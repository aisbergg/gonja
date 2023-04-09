package exec

import (
	"strings"

	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

func getBlocks(tpl *parse.TemplateNode) map[string]*parse.WrapperNode {
	if tpl == nil {
		return map[string]*parse.WrapperNode{}
	}
	blocks := getBlocks(tpl.Parent)
	for name, wrapper := range tpl.Blocks {
		blocks[name] = wrapper
	}
	return blocks
}

func Self(r *Renderer) map[string]func() (string, error) {
	blocks := map[string]func() (string, error){}
	for name, block := range getBlocks(r.Root) {
		blocks[name] = func() (string, error) {
			sub := r.Inherit()
			var out strings.Builder
			sub.Out = &out
			if err := sub.ExecuteWrapper(block); err != nil {
				return "", err
			}
			return out.String(), nil
		}
	}
	return blocks
}
