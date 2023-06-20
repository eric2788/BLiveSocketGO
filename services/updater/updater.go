package updater

import (
	"context"
	"encoding/json"
	"github.com/rogpeppe/go-internal/semver"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

const (
	repoUrl = "https://api.github.com/repos/eric2788/biligo-live-ws/releases/latest"
)

var VersionTag string

var (
	log = logrus.WithField("service", "updater")
	ctx = context.Background()
)

func StartUpdater() {
	log.Info("已啟動更新檢查器")
	tick := time.NewTicker(time.Hour * 24)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			log.Info("正在檢查更新")
			if resp, err := checkForUpdates(); err != nil {
				log.Warnf("檢查更新時出現錯誤: %v", err)
			} else {
				if semver.Compare(resp.TagName, VersionTag) > 0 && !resp.Prerelease {
					log.Infof("有可用新版本: %s", resp.TagName)
					log.Infof("請自行到 %v 下載。", resp.HtmlUrl)
				} else {
					log.Infof("你目前已是最新版本。")
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func checkForUpdates() (*ReleaseLatestResp, error) {
	res, err := http.Get(repoUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var resp = &ReleaseLatestResp{}
	err = json.Unmarshal(b, resp)
	return resp, err
}
