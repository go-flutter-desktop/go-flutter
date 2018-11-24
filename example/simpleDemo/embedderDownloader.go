package main

import (
    "io"
    "fmt"
    "log"
    "os/exec"
    "runtime"
    "regexp"
    "net/http"
    "os"
    "io/ioutil"
    "encoding/json"
)

// Function to download file with given path and url.
func DownloadFile(filepath string, url string) error {

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

    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    if err != nil {
        return err
    }

    return nil
}

func main() {
    // Execute flutter command to retrieve the version
	out, err := exec.Command("flutter","--version").Output()
    if err != nil {
        log.Fatal(err)
    }
    
    // DEBUG
    // fmt.Printf("The os is %s\n",platform)

    re := regexp.MustCompile(`Engine • revision (\w{10})`)
    shortRevision := re.FindStringSubmatch(string(out))[1]

    url := fmt.Sprintf("https://api.github.com/search/commits?q=%s", shortRevision)
    
    // DEBUG
    //fmt.Printf("The url is %s\n",url)

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
    myStruct := struct {
		IncompleteResults bool `json:"incomplete_results"`
		Items             []struct {
			Sha string `json:"sha"`
			URL string `json:"url"`
		} `json:"items"`
		TotalCount int `json:"total_count"`
	}{}

	err2 := json.Unmarshal(body, &myStruct)
    if err2 != nil {
        // handle err
        log.Fatal(err2)
    }

	var platform = "undefined"
    var downloadUrl = ""
	
    // Retrieve the OS and set variable to retrieve correct flutter embedder
    switch runtime.GOOS {
    case "darwin":
        platform = "darwin-x64"
        downloadUrl = fmt.Sprintf("https://storage.googleapis.com/flutter_infra/flutter/%s/%s/FlutterEmbedder.framework.zip", myStruct.Items[0].Sha, platform)
    case "linux":
        platform = "linux-x64"
        downloadUrl = fmt.Sprintf("https://storage.googleapis.com/flutter_infra/flutter/%s/%s/%s-embedder", myStruct.Items[0].Sha, platform, platform)

    case "windows":
        platform = "windows-x64"
        downloadUrl = fmt.Sprintf("https://storage.googleapis.com/flutter_infra/flutter/%s/%s/%s-embedder", myStruct.Items[0].Sha, platform, platform)

    default:
        log.Fatal("OS not supported")
    }

    err3 := DownloadFile(".build/temp.zip", downloadUrl)
    if err3 != nil {
        log.Fatal(err3)
    } else{
        fmt.Printf("Downloaded embedder for %s platform, matching version : %s\n", platform, myStruct.Items[0].Sha)
    }
}