package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func createSymLink(symlink string, file string) {

	os.Remove(symlink)

	err := os.Symlink(file, symlink)
	if err != nil {
		log.Fatal(err)
	}
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			if err != nil {
				return filenames, err
			}

		}
	}
	return filenames, nil
}

func askForConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Would you like to overwrite the previously downloaded engine [Y/n] : ")

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" || response == "" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

// Function to prind download percent completion
func printDownloadPercent(done chan int64, path string, total int64) {

	var stop = false

	for {
		select {
		case <-done:
			stop = true
		default:

			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}

			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			size := fi.Size()

			if size == 0 {
				size = 1
			}

			var percent = float64(size) / float64(total) * 100

			// We use `\033[2K\r` to avoid carriage return, it will print above previous.
			fmt.Printf("\033[2K\r %.0f %% / 100 %%", percent)
		}

		if stop {
			break
		}

		time.Sleep(time.Second)
	}
}

// Function to download file with given path and url.
func downloadFile(filepath string, url string) error {

	// Print download url in case user needs it.
	fmt.Printf("Downloading file from\n '%s'\n to '%s'\n\n", url, filepath)

	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		if !askForConfirmation() {
			fmt.Printf("Leaving.\n")
			os.Exit(0)
		}
	}

	start := time.Now()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))

	done := make(chan int64)

	go printDownloadPercent(done, filepath, int64(size))

	// Write the body to file
	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	done <- n

	elapsed := time.Since(start)
	log.Printf("\033[2K\rDownload completed in %s", elapsed)

	return nil
}

func main() {
	// Support flag
	chinaPtr := flag.Bool("china", false, "Whether or not installation is in China")
	flag.Parse()
	var targetedDomain = ""

	// If flag china is passse, targeted domain is changed (China partially blocking google)
	if *chinaPtr {
		targetedDomain = "https://storage.flutter-io.cn"
	} else {
		targetedDomain = "https://storage.googleapis.com"
	}

	// Execute flutter command to retrieve the version
	out, err := exec.Command("flutter", "--version").Output()
	if err != nil {
		log.Fatal(err)
	}

	// Get working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile(`Engine • revision (\w{10})`)
	shortRevision := re.FindStringSubmatch(string(out))[1]

	url := fmt.Sprintf("https://api.github.com/search/commits?q=%s", shortRevision)

	// This part is used to retrieve the full hash
	req, err := http.NewRequest("GET", os.ExpandEnv(url), nil)
	if err != nil {
		// handle err
		log.Fatal(err)
	}
	req.Header.Set("Accept", "application/vnd.github.cloak-preview")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	// We define a struct to build JSON object from the response
	hashResponse := struct {
		Items []struct {
			Sha string `json:"sha"`
		} `json:"items"`
	}{}

	err2 := json.Unmarshal(body, &hashResponse)
	if err2 != nil {
		// handle err
		log.Fatal(err2)
	}

	var platform = "undefined"
	var downloadShareLibraryURL = ""
	var endMessage = ""

	// Retrieve the OS and set variable to retrieve correct flutter embedder
	switch runtime.GOOS {
	case "darwin":
		platform = "darwin-x64"
		downloadShareLibraryURL = fmt.Sprintf(targetedDomain+"/flutter_infra/flutter/%s/%s/FlutterEmbedder.framework.zip", hashResponse.Items[0].Sha, platform)
		endMessage = "export CGO_LDFLAGS=\"-F${PWD} -Wl,-rpath,@executable_path\""

	case "linux":
		platform = "linux-x64"
		downloadShareLibraryURL = fmt.Sprintf(targetedDomain+"/flutter_infra/flutter/%s/%s/%s-embedder", hashResponse.Items[0].Sha, platform, platform)
		endMessage = "export CGO_LDFLAGS=\"-L${PWD}\""

	case "windows":
		platform = "windows-x64"
		downloadShareLibraryURL = fmt.Sprintf(targetedDomain+"/flutter_infra/flutter/%s/%s/%s-embedder", hashResponse.Items[0].Sha, platform, platform)
		endMessage = "set CGO_LDFLAGS=-L%cd%"

	default:
		log.Fatal("OS not supported")
	}

	downloadIcudtlURL := fmt.Sprintf(targetedDomain+"/flutter_infra/flutter/%s/%s/artifacts.zip", hashResponse.Items[0].Sha, platform)

	err3 := downloadFile(dir+"/.build/temp.zip", downloadShareLibraryURL)
	if err3 != nil {
		log.Fatal(err3)
	} else {
		fmt.Printf("Downloaded embedder for %s platform, matching version : %s\n", platform, hashResponse.Items[0].Sha)
	}

	err4 := downloadFile(dir+"/.build/artifacts.zip", downloadIcudtlURL)
	if err != nil {
		log.Fatal(err4)
	} else {
		fmt.Printf("Downloaded artifact for %s platform.\n", platform)
	}

	_, err = unzip(".build/temp.zip", dir+"/.build/")
	if err != nil {
		log.Fatal(err)
	}

	_, err = unzip(".build/artifacts.zip", dir+"/.build/artifacts/")
	if err != nil {
		log.Fatal(err)
	}

	err = os.Rename(".build/artifacts/icudtl.dat", dir+"/icudtl.dat")
	if err != nil {
		log.Fatal(err)
	}

	switch platform {
	case "darwin-x64":
		_, err = unzip(dir+"/.build/FlutterEmbedder.framework.zip", dir+"/.build/FlutterEmbedder.framework/")
		if err != nil {
			log.Fatal(err)
		}

		os.RemoveAll(dir + "/FlutterEmbedder.framework")

		err := os.Rename(dir+"/.build/FlutterEmbedder.framework/", dir+"/FlutterEmbedder.framework/")
		if err != nil {
			log.Fatal(err)
		}

		createSymLink(dir+"/FlutterEmbedder.framework/Versions/Current", dir+"/FlutterEmbedder.framework/Versions/A")

		createSymLink(dir+"/FlutterEmbedder.framework/FlutterEmbedder", dir+"/FlutterEmbedder.framework/Versions/Current/FlutterEmbedder")

		createSymLink(dir+"/FlutterEmbedder.framework/Headers", dir+"/FlutterEmbedder.framework/Versions/Current/Headers")

		createSymLink(dir+"/FlutterEmbedder.framework/Modules", dir+"/FlutterEmbedder.framework/Versions/Current/Modules")

		createSymLink(dir+"/FlutterEmbedder.framework/Resources", dir+"/FlutterEmbedder.framework/Versions/Current/Resources")

	case "linux-x64":
		err := os.Rename(".build/libflutter_engine.so", dir+"/libflutter_engine.so")
		if err != nil {
			log.Fatal(err)
		}

	case "windows-x64":
		err := os.Rename(".build/flutter_engine.dll", dir+"/flutter_engine.dll")
		if err != nil {
			log.Fatal(err)
		}

	}
	fmt.Printf("Unzipped files and moved them to correct repository.\n")

	fmt.Printf("\nTo let know the CGO compiler where to look for the share library, Please run:\n\t")
	fmt.Printf("%s\n", endMessage)

	fmt.Printf("Done.\n")
}
