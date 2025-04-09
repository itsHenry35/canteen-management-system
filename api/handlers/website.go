package handlers

import (
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
func GetWebsiteInfo(w http.ResponseWriter, r *http.Request) {
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
