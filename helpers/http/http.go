package http

import (
	"os"
	"os/exec"
)

func HttpFetchFileProgress(url string, outFile string, stdout bool) (err error) {
	/*
		out, err := os.Create(outFile)
		defer out.Close()
		if err != nil {
			return
		}
		response, err := http.Get(url)
		if err != nil {
			return
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			err = errors.New("Server responded: " + response.Status)
			return
		}
		pb := progress.NewProgress(os.Stderr, response.ContentLength, stdout)

		io.Copy(pb, response.Body)
		if stdout {
			fmt.Println()
		}
	*/
	cmd := exec.Command("curl", "-L", url, "-o", outFile, "-#")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
