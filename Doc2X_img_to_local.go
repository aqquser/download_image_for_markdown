// 编译命令  
// go build -o Doc2X_img_to_local_golang.exe   Doc2X_img_to_local.go
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ImageDownloader struct {
	saveDir string
}

func NewImageDownloader(saveDir string) *ImageDownloader {
	return &ImageDownloader{
		saveDir: saveDir,
	}
}

func (id *ImageDownloader) downloadImage(url string, savePath string) bool {
	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("下载图片失败 %s: %v\n", url, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	out, err := os.Create(savePath)
	if err != nil {
		fmt.Printf("创建文件失败 %s: %v\n", savePath, err)
		return false
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("保存图片失败 %s: %v\n", savePath, err)
		return false
	}

	return true
}

func (id *ImageDownloader) processMarkdownFile(filePath string) {
	fmt.Printf("处理文件: %s\n", filePath)

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("读取文件失败 %s: %v\n", filePath, err)
		return
	}

	imgPattern := regexp.MustCompile(`<img src="(.*?)"/>`)
	modified := false
	imageCount := 0
	noUpdatePrinted := false

	newContent := imgPattern.ReplaceAllStringFunc(string(content), func(match string) string {
		imageCount++
		
		// 提取URL
		submatches := imgPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		originalURL := submatches[1]

		// 检查URL是否包含noedgeai.com
		if !strings.Contains(originalURL, "noedgeai.com") {
			if !noUpdatePrinted {
				fmt.Println("所有图片不包含noedgeai.com,文件无需更新")
				noUpdatePrinted = true
			}
			return match
		}

		// 获取文件扩展名
		fileExt := filepath.Ext(strings.Split(originalURL, "?")[0])
		if fileExt == "" {
			fileExt = ".jpg"
		}

		// 构建新文件名
		baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		newFilename := fmt.Sprintf("%s-%d%s", baseName, imageCount, fileExt)
		savePath := filepath.Join(id.saveDir, newFilename)

		// 下载图片
		if id.downloadImage(originalURL, savePath) {
			modified = true
			return fmt.Sprintf(`<img src="%s"/>`, savePath)
		}

		return match
	})

	if modified {
		err = ioutil.WriteFile(filePath, []byte(newContent), 0644)
		if err != nil {
			fmt.Printf("写入文件失败 %s: %v\n", filePath, err)
			return
		}
		fmt.Println("文件已更新")
	} else if !noUpdatePrinted {
		fmt.Println("文件无需更新")
	}
}

func main() {
	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("获取程序路径失败: %v\n", err)
		return
	}
	markdownDir := filepath.Dir(execPath)
	
	// 获取当前用户名和默认保存路径
	currentUser, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("获取用户目录失败: %v\n", err)
		return
	}
	defaultSaveDir := filepath.Join(currentUser, "Pictures", "markdown图片")
	
	// 获取用户输入的保存路径
	fmt.Printf("请输入图片保存路径 (直接回车使用默认路径: %s): \n", defaultSaveDir)
	var saveDir string
	fmt.Scanln(&saveDir)
	
	// 如果用户直接按回车，使用默认路径
	if saveDir == "" {
		saveDir = defaultSaveDir
	}
	
	// 验证路径是否存在
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		fmt.Printf("路径 %s 不存在，是否创建? (y/n): ", saveDir)
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) == "y" {
			err = os.MkdirAll(saveDir, 0755)
			if err != nil {
				fmt.Printf("创建目录失败: %v\n", err)
				fmt.Println("\n按回车键退出...")
				fmt.Scanln()
				return
			}
		} else {
			fmt.Println("程序退出")
			fmt.Println("\n按回车键退出...")
			fmt.Scanln()
			return
		}
	}

	downloader := NewImageDownloader(saveDir)

	// 遍历目录中的所有markdown文件
	err = filepath.Walk(markdownDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			downloader.processMarkdownFile(path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("遍历目录失败: %v\n", err)
	}
	
	fmt.Println("\n处理完成！按回车键退出...")
	fmt.Scanln()
}
