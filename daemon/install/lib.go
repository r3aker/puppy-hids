package install

import (
	"github.com/thonsun/puppy-hids/daemon/common"
	"github.com/thonsun/puppy-hids/daemon/log"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

//下载agent 文件到指定位置，并更改权限
func downFile(url string, savepath string) error {
	request, _ := http.NewRequest("GET", url, nil)
	request.Close = true
	if res, err := common.HTTPClient.Do(request); err == nil {
		defer res.Body.Close()
		file, err := os.Create(savepath)
		if err != nil {
			return err
		}
		io.Copy(file, res.Body)
		file.Close()
		os.Chmod(savepath, 0750) // 配置运行权限

		fileInfo, err := os.Stat(savepath)
		log.Debug("agent download status:%v %v",res.ContentLength, fileInfo.Size())
		if err != nil || fileInfo.Size() != res.ContentLength {
			log.Debug("File download error: %v", err.Error())
			return errors.New("downfile error")
		}
	} else {
		return err
	}
	return nil
}

//复制自身
func copyMe(installPath string) (err error) {
	var dstName string
	dstName = installPath + "daemon"
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}
	mepath, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	if mepath == dstName {
		// 相同目录不用复制
		return nil
	}
	src, err := os.Open(mepath)// 找到自身文件所在位置
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	return nil
}

// DownAgent 下载agent到指定安装目录
func DownAgent(ip string, agentPath string, arch string) error {
	var err error
	url := fmt.Sprintf("%s://%s/json/download", common.Proto, ip)
	log.Debug("agent download:%s",url)

	// Agent 下载检查和重试, 重试三次，功能性考虑
	times := 3
	for {
		err = downFile(url, agentPath)
		// 检查文件hash是否匹配
		if err == nil {
			mstr, _ := FileMD5String(agentPath)
			log.Debug("Agent file MD5:%v", mstr)
			if CheckAgentHash(mstr, ip, arch) {
				log.Debug("Agent download finished, hash check passed")
				return nil
			} else {
				log.Error("Agent is broken, retry the downloader again")
			}
		}
		if times--; times == 0 {
			break
		}
	}

	return errors.New("Agent Download Error")
}

// FileMD5String 获取文件MD5
func FileMD5String(filePath string) (MD5String string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	md5h := md5.New()
	io.Copy(md5h, file)
	return fmt.Sprintf("%x", md5h.Sum([]byte(""))), nil
}

// CheckAgentHash 检查Agent的哈希值是否匹配
func CheckAgentHash(fileHash string, ip string, arch string) (is bool) {
	checkURL := fmt.Sprintf(
		"%s://%s/json/check?md5=%s",
		common.Proto, ip, fileHash,
	)
	res, err := common.HTTPClient.Get(checkURL)
	if err != nil {
		return false
	}
	resp, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false
	}
	var result string
	err = json.Unmarshal(resp,&result)
	if err != nil {
		return false
	}
	log.Debug("recieve from web md5:%s",result)
	return "1" == string(result)
}