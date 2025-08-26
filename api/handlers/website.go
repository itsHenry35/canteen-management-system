package handlers

import (
	"fmt"
	"net/http"

	"github.com/itsHenry35/canteen-management-system/config"
	"github.com/itsHenry35/canteen-management-system/utils"
)

// WebsiteInfoResponse 网站信息响应
type WebsiteInfoResponse struct {
	Name           string `json:"name"`             // 网站名称
	ICPBeian       string `json:"icp_beian"`        // ICP备案信息
	PublicSecBeian string `json:"public_sec_beian"` // 公安部备案信息
	DingTalkCorpID string `json:"dingtalk_corp_id"` // 钉钉企业ID
	Domain         string `json:"domain"`           // 网站域名
}

// GetWebsiteInfo 获取网站信息
func GetWebsiteInfo(w http.ResponseWriter, _ *http.Request) {
	// 获取配置
	cfg := config.Get()

	// 构建响应
	resp := WebsiteInfoResponse{
		Name:           cfg.Website.Name,
		ICPBeian:       cfg.Website.ICPBeian,
		PublicSecBeian: cfg.Website.PublicSecBeian,
		DingTalkCorpID: cfg.DingTalk.CorpID,
		Domain:         cfg.Website.Domain,
	}

	// 返回响应
	utils.ResponseOK(w, resp)
}

// GetManifest 获取 manifest.webmanifest 内容
func GetManifest(w http.ResponseWriter, _ *http.Request) {
	// 获取配置
	cfg := config.Get()

	// 构建manifest
	resp := fmt.Sprintf(`
	{
		"short_name": "%s",
		"name": "%s",
		"icons": [
			{
			"src": "logo192.png",
			"type": "image/png",
			"sizes": "192x192"
			},
			{
			"src": "logo512.png",
			"type": "image/png",
			"sizes": "512x512"
			},
			{
			"src": "favicon.svg",
			"sizes": "any",
			"type": "image/svg+xml"
			}
		],
		"start_url": ".",
		"display": "standalone",
		"theme_color": "#001529",
		"background_color": "#f5f5f5"
		}
	`, cfg.Website.Name, cfg.Website.Name)

	// 返回manifest
	w.Header().Set("Content-Type", "application/manifest+json")
	w.Write([]byte(resp))
}
