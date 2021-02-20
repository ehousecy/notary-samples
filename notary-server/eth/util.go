package eth

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

func home() (string, error){
	cu, err := user.Current()
	if err != nil {
		switch runtime.GOOS {
		case "windows":
			return homeWindows()
		default:
			return homeUnix()
		}
	}
	return cu.HomeDir, nil
}

func homeUnix()(string, error)  {
	if home := os.Getenv("HOME"); home != ""{
		return home, nil
	}
	var stdout []byte
	_, err := exec.Command("sh", "-c", "eval echo ~$USER").Stdout.Write(stdout)
	if err != nil {
		return "", err
	}
	result := strings.TrimSpace(string(stdout))
	if result == ""{
		return "", errors.New("empty home directory")
	}
	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := filepath.Join(drive, path)
	if drive == "" || path == ""{
		home = os.Getenv("USERPROFILE")
	}
	if home == ""{
		return "", errors.New("home drive, home path, and user profile are blank")
	}
	return home, nil
}

//get the configuration home directory path
func ConfigHome() string {
	userHome,_ := home()
	configPath := filepath.Join(userHome, ".notary-sample")

	//create home director if director does not exist
	fileInfo, err := os.Stat(configPath)
	if err != nil || !fileInfo.IsDir() {
		err = os.MkdirAll(configPath, os.ModePerm)
		if err != nil {
			fmt.Println("Error occurred when creating directory:",err.Error())
		}
	}
	return configPath
}