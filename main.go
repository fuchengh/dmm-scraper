package main

import (
	"fmt"
	"image"
	"os"
	"path"
	"path/filepath"
	"strings"

	"dmm-scraper/pkg/config"
	"dmm-scraper/pkg/img"
	"dmm-scraper/pkg/logger"
	"dmm-scraper/pkg/metadata"
	"dmm-scraper/pkg/scraper"

	"github.com/imroc/req/v3"
	"github.com/schollz/progressbar/v3"
)

var (
	ratio = 1.42
)

type VideoFile struct {
	Entry os.DirEntry
	Path string
}

func IsVideo(entry os.DirEntry) bool {
	ext := filepath.Ext(entry.Name())
	switch ext {
	case ".mp4", ".avi", ".mov", ".mkv", ".flv", ".wmv":
		return true
	default:
		return false
	}
}

func MyProgress(l logger.Logger, sType, filename string) func(info req.DownloadInfo) {
	bar := progressbar.NewOptions64(100,
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription(fmt.Sprintf("[light_blue]Download[reset] %s", filename)),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
	)

	return func(info req.DownloadInfo) {
		if info.Response.StatusCode != 200 {
			return
		}

		current := info.DownloadedSize
		total := info.Response.ContentLength

		if total <= 0 {
			return
		}

		percentage := float64(current * 100 / total) 
		bar.Set64(int64(percentage))
	}
}

func collectVideoFiles(root string, ignored []string) ([]VideoFile, error) {
	var files []VideoFile
	ignored = append(ignored, "Failed") // auto-ignore "Failed" directory

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip ignored directories
		if entry.IsDir() {
			for _, ig := range ignored {
				if strings.EqualFold(entry.Name(), ig) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if IsVideo(entry) {
			files = append(files, VideoFile{
				Entry: entry,
				Path:  path,
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return files, nil
}


func main() {
	var err error
	log := logger.New()

	conf, err := config.NewLoader().LoadFile("config")
	if err != nil {
		log.Errorf("Error reading config file, %s", err)
		log.Warnf("Loading default config")
		conf = config.Default()
	}

	scraper.Setup(conf)

	files, err := collectVideoFiles(conf.Input.Path, conf.Input.Ignored)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("============== Found %d video files ==============", len(files))

	for _, f := range files {
		log.Infof("-------- Processing movie ID: %s --------", f.Entry.Name())
		ext := filepath.Ext(f.Entry.Name())
		name := strings.TrimSuffix(f.Entry.Name(), ext)
		failedPath := path.Join(conf.Input.Path, "Failed", f.Entry.Name())

		// 用正则处理文件名
		if query, scrapers := scraper.GetQuery(name); query != "" {

			for _, s := range scrapers {
				log.Debugf("%s capturing query: %s", s.GetType(), query)

				// fetch
				err = s.FetchDoc(query)
				if err != nil {
					log.Error(err)
					MoveFailedFile(f.Path, failedPath, conf.Input.MoveFail)
					continue
				}

				if s.GetNumber() == "" {
					log.Errorf("%s get num empty", s.GetType())
					MoveFailedFile(f.Path, failedPath, conf.Input.MoveFail)
					continue
				}

				num := s.GetFormatNumber()
				log.Infof("%s parsing movie ID %s ", s.GetType(), num)

				// mkdir
				outputPath := scraper.GetOutputPath(s, conf.Output.Path)
				log.Debugf("%s making output path: %s", s.GetType(), outputPath)
				err = os.MkdirAll(outputPath, 0700)
				if err != nil && !os.IsExist(err) {
					log.Error(err)
					MoveFailedFile(f.Path, failedPath, conf.Input.MoveFail)
					break
				}

				// build nfo
				movieNfo := metadata.NewMovieNfo(s)
				cover := fmt.Sprintf("%s-fanart.jpg", num)
				poster := fmt.Sprintf("%s-poster.jpg", num)
				// movieNfo.SetTitle(num)

				// Get fanart and poster
				coverPath := path.Join(outputPath, cover)
				posterPath := path.Join(outputPath, poster)
				err = scraper.Download(s.GetCover(), coverPath, MyProgress(log, s.GetType(), cover))
				if err != nil {
					log.Error(err)
					MoveFailedFile(f.Path, failedPath, conf.Input.MoveFail)
					break
				}
				// Get poster
				// first, check if we can get poster from website
				err = scraper.Download(s.GetPoster(), posterPath, MyProgress(log, s.GetType(), poster))
				if err != nil || !img.ValidPosterProportion(posterPath) {
					// fallback to crop from cover
					log.Warnf("%s failed to fetch cover, crop from fanart", s.GetType())
					imgOperation := img.NewOperation()
					// calculate posterWidth based on cover width
					posterWidth, _ := getPosterWidth(coverPath, ratio)
					err = imgOperation.CropAndSave(coverPath, posterPath, posterWidth, 0)
					if err != nil {
						log.Errorf("Failed to get poster: %v", err)
						MoveFailedFile(f.Path, failedPath, conf.Input.MoveFail)
						break
					}
				}

				nfo := path.Join(outputPath, fmt.Sprintf("%s.nfo", num))
				log.Debugf("%s writing nfo file: %s", s.GetType(), nfo)
				err = movieNfo.Save(nfo)
				if err != nil {
					log.Error(err)
					MoveFailedFile(f.Path, failedPath, conf.Input.MoveFail)
					break
				}

				// do not move if file is already exist
				fileExist := path.Join(outputPath, f.Entry.Name())
				if _, err := os.Stat(fileExist); err == nil {
					log.Infof("%s file already exist: %s, skip moving", s.GetType(), fileExist)
					MoveFailedFile(f.Path, failedPath, conf.Input.MoveFail)
					break
				}

				log.Debugf("%s moving video file to: %s", s.GetType(), outputPath)
				err = MoveFile(f.Path, outputPath, num, 1)
				if err != nil {
					log.Error(err)
					MoveFailedFile(f.Path, failedPath, conf.Input.MoveFail)
					break
				}

				break
			}
		}
	}
}

func MoveFile(oldPath, outputPath, num string, index int) error {
	var filename string
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return err
	}

	if index != 1 {
		filename = fmt.Sprintf("%s-cd%d%s", num, index, filepath.Ext(oldPath))
	} else {
		filename = fmt.Sprintf("%s%s", num, filepath.Ext(oldPath))
	}
	newPath := path.Join(outputPath, filename)
	if _, err := os.Stat(newPath); err == nil {
		index += 1
		return MoveFile(oldPath, outputPath, num, index)
	}
	return os.Rename(oldPath, newPath)
}

func getPosterWidth(fanartPath string, ratio float64) (height int, width int) {
	if reader, err := os.Open(fanartPath); err == nil {
		defer reader.Close()
		im, _, err := image.Decode(reader)
		if err != nil {
			return 378, 0
		}

		var posterH, posterW int

		fanartH, fanartW := im.Bounds().Dy(), im.Bounds().Dx()
		if float64(fanartH)/float64(fanartW) < ratio {
			posterW = int(float64(fanartH) / ratio)
			posterH = fanartH
		} else {
			posterW = fanartW
			posterH = int(float64(fanartW) * ratio)
		}
		return posterW, posterH
	}
	return 378, 0
}

func MoveFailedFile(sourcePath, destPath string, moveFail bool) error {
	if !moveFail {
		return nil
	}
    // Get the directory part of the destination path
    destDir := filepath.Dir(destPath)

    // Check if destination directory exists, create if not exist
    if _, err := os.Stat(destDir); os.IsNotExist(err) {
        err = os.MkdirAll(destDir, os.ModePerm)
        if err != nil {
            return err
        }
    }

    // Move the file
    err := os.Rename(sourcePath, destPath)
    if err != nil {
        return err
    }
    return nil
}