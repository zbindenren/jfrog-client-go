package services

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/httpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"net/http"
	"path"
)

type PromoteService struct {
	client     *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
	DryRun     bool
}

func NewPromotionService(client *httpclient.HttpClient) *PromoteService {
	return &PromoteService{client: client}
}

func (ps *PromoteService) isDryRun() bool {
	return ps.DryRun
}

func (ps *PromoteService) BuildPromote(promotionParams PromotionParams) error {
	message := "Promoting build..."
	if ps.DryRun == true {
		message = "[Dry run] " + message
	}
	log.Info(message)

	promoteUrl := ps.ArtDetails.GetUrl()
	restApi := path.Join("api/build/promote/", promotionParams.GetBuildName(), promotionParams.GetBuildNumber())
	requestFullUrl, err := utils.BuildArtifactoryUrl(promoteUrl, restApi, make(map[string]string))
	if err != nil {
		return err
	}

	data := BuildPromotionBody{
		Status:              promotionParams.GetStatus(),
		Comment:             promotionParams.GetComment(),
		Copy:                promotionParams.IsCopy(),
		IncludeDependencies: promotionParams.IsIncludeDependencies(),
		SourceRepo:          promotionParams.GetSourceRepo(),
		TargetRepo:          promotionParams.GetTargetRepo(),
		DryRun:              ps.isDryRun()}
	requestContent, err := json.Marshal(data)
	if err != nil {
		return errorutils.CheckError(err)
	}

	httpClientsDetails := ps.ArtDetails.CreateHttpClientDetails()
	utils.SetContentType("application/vnd.org.jfrog.artifactory.build.PromotionRequest+json", &httpClientsDetails.Headers)

	resp, body, err := ps.client.SendPost(requestFullUrl, requestContent, httpClientsDetails)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Artifactory response:", resp.Status)
	log.Info("Promoted build", promotionParams.GetBuildName()+"/"+promotionParams.GetBuildNumber(), "to:", promotionParams.GetTargetRepo(), "repository.")
	return nil
}

type BuildPromotionBody struct {
	Comment             string `json:"comment,omitempty"`
	SourceRepo          string `json:"sourceRepo,omitempty"`
	TargetRepo          string `json:"targetRepo,omitempty"`
	Status              string `json:"status,omitempty"`
	IncludeDependencies bool   `json:"dependencies,omitempty"`
	Copy                bool   `json:"copy,omitempty"`
	DryRun              bool   `json:"dryRun,omitempty"`
}

type PromotionParams struct {
	BuildName           string
	BuildNumber         string
	TargetRepo          string
	Status              string
	Comment             string
	Copy                bool
	IncludeDependencies bool
	SourceRepo          string
}

func (bp *PromotionParams) GetBuildName() string {
	return bp.BuildName
}

func (bp *PromotionParams) GetBuildNumber() string {
	return bp.BuildNumber
}

func (bp *PromotionParams) GetTargetRepo() string {
	return bp.TargetRepo
}

func (bp *PromotionParams) GetStatus() string {
	return bp.Status
}

func (bp *PromotionParams) GetComment() string {
	return bp.Comment
}

func (bp *PromotionParams) IsCopy() bool {
	return bp.Copy
}

func (bp *PromotionParams) IsIncludeDependencies() bool {
	return bp.IncludeDependencies
}

func (bp *PromotionParams) GetSourceRepo() string {
	return bp.SourceRepo
}

func NewPromotionParams() PromotionParams {
	return PromotionParams{}
}
