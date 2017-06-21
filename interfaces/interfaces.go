package interfaces

import (
	"github.com/LSvKing/centipede/items"
	"github.com/LSvKing/centipede/pipeline"
)

type Spider interface {
	Add(url string)
	AddAll(urls []string)
	Run()
	AddRequest(url string, callback string) *Spider
	AddReq(url string, callback string) *Spider
	AddRule(ruleTree *items.RuleTree) *Spider
	AddStarUrls(starUrls []string) *Spider
	GetRule() *items.RuleTree
	AddPipeline(pipeline pipeline.Pipeline) *Spider
}
