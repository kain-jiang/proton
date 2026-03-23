package slb

import (
	"bufio"
	"fmt"
	"sort"
	"strings"
)

type keepalivedNode interface {
	render(*strings.Builder, int)
}

type keepalivedKey struct {
	Name  string
	Value string
}

func (k *keepalivedKey) render(b *strings.Builder, indent int) {
	writeIndent(b, indent)
	b.WriteString(k.Name)
	if k.Value != "" {
		b.WriteByte(' ')
		b.WriteString(k.Value)
	}
	b.WriteByte('\n')
}

type keepalivedBlock struct {
	Kind  string
	Name  string
	Items []keepalivedNode
}

func (blk *keepalivedBlock) render(b *strings.Builder, indent int) {
	writeIndent(b, indent)
	b.WriteString(blk.Kind)
	if blk.Name != "" {
		b.WriteByte(' ')
		b.WriteString(blk.Name)
	}
	b.WriteString(" {\n")
	for _, item := range blk.Items {
		item.render(b, indent+1)
	}
	writeIndent(b, indent)
	b.WriteString("}\n\n")
}

func (blk *keepalivedBlock) key(name string) *keepalivedKey {
	for _, item := range blk.Items {
		key, ok := item.(*keepalivedKey)
		if ok && key.Name == name {
			return key
		}
	}
	return nil
}

func (blk *keepalivedBlock) setKey(name, value string) {
	if key := blk.key(name); key != nil {
		key.Value = value
		return
	}
	blk.Items = append(blk.Items, &keepalivedKey{Name: name, Value: value})
}

func (blk *keepalivedBlock) child(kind string) *keepalivedBlock {
	for _, item := range blk.Items {
		child, ok := item.(*keepalivedBlock)
		if ok && child.Kind == kind {
			return child
		}
	}
	return nil
}

func (blk *keepalivedBlock) setChild(newBlock *keepalivedBlock) {
	for i, item := range blk.Items {
		child, ok := item.(*keepalivedBlock)
		if ok && child.Kind == newBlock.Kind {
			blk.Items[i] = newBlock
			return
		}
	}
	blk.Items = append(blk.Items, newBlock)
}

type keepalivedConf struct {
	Items []keepalivedNode
}

func parseKeepalivedConfig(data []byte) (*keepalivedConf, error) {
	conf := &keepalivedConf{}
	var stack []*keepalivedBlock

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "", strings.HasPrefix(line, "#"):
			continue
		case strings.HasSuffix(line, "{"):
			header := strings.TrimSpace(strings.TrimSuffix(line, "{"))
			kind, name := splitKeepalivedDirective(header)
			block := &keepalivedBlock{Kind: kind, Name: name}
			if len(stack) == 0 {
				conf.Items = append(conf.Items, block)
			} else {
				stack[len(stack)-1].Items = append(stack[len(stack)-1].Items, block)
			}
			stack = append(stack, block)
		case line == "}":
			if len(stack) == 0 {
				return nil, fmt.Errorf("unexpected keepalived block end")
			}
			stack = stack[:len(stack)-1]
		default:
			name, value := splitKeepalivedDirective(line)
			key := &keepalivedKey{Name: name, Value: value}
			if len(stack) == 0 {
				conf.Items = append(conf.Items, key)
			} else {
				stack[len(stack)-1].Items = append(stack[len(stack)-1].Items, key)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(stack) != 0 {
		return nil, fmt.Errorf("unexpected keepalived block nesting")
	}
	return conf, nil
}

func splitKeepalivedDirective(line string) (name, value string) {
	line = strings.TrimSpace(line)
	idx := strings.IndexAny(line, " \t")
	if idx < 0 {
		return line, ""
	}
	return line[:idx], strings.TrimSpace(line[idx+1:])
}

func (conf *keepalivedConf) render() []byte {
	var b strings.Builder
	for _, item := range conf.orderedItems() {
		item.render(&b, 0)
	}
	return []byte(b.String())
}

func (conf *keepalivedConf) orderedItems() []keepalivedNode {
	ordered := make([]keepalivedNode, 0, len(conf.Items))
	appendBlocks := func(kind string) {
		for _, item := range conf.Items {
			block, ok := item.(*keepalivedBlock)
			if ok && block.Kind == kind {
				ordered = append(ordered, block)
			}
		}
	}
	appendBlocks("global_defs")
	appendBlocks("vrrp_script")
	appendBlocks("vrrp_instance")
	appendBlocks("virtual_server")

	for _, item := range conf.Items {
		block, ok := item.(*keepalivedBlock)
		if ok {
			switch block.Kind {
			case "global_defs", "vrrp_script", "vrrp_instance", "virtual_server":
				continue
			}
		}
		ordered = append(ordered, item)
	}
	return ordered
}

func (conf *keepalivedConf) blocks(kind string) []*keepalivedBlock {
	var result []*keepalivedBlock
	for _, item := range conf.Items {
		block, ok := item.(*keepalivedBlock)
		if ok && block.Kind == kind {
			result = append(result, block)
		}
	}
	return result
}

func (conf *keepalivedConf) block(kind, name string) *keepalivedBlock {
	for _, item := range conf.Items {
		block, ok := item.(*keepalivedBlock)
		if ok && block.Kind == kind && block.Name == name {
			return block
		}
	}
	return nil
}

func (conf *keepalivedConf) upsertBlock(block *keepalivedBlock) {
	for i, item := range conf.Items {
		existing, ok := item.(*keepalivedBlock)
		if ok && existing.Kind == block.Kind && existing.Name == block.Name {
			conf.Items[i] = block
			return
		}
	}
	conf.Items = append(conf.Items, block)
}

func (conf *keepalivedConf) ensureGlobalDefs() *keepalivedBlock {
	if block := conf.block("global_defs", ""); block != nil {
		return block
	}
	block := &keepalivedBlock{Kind: "global_defs"}
	conf.Items = append(conf.Items, block)
	return block
}

func sortedKeys[V any](in map[string]V) []string {
	keys := make([]string, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
