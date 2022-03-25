package patcher

import (
	"bytes"
	"fmt"
	"sort"
)

type patch struct {
	offset int
	length int
	data   []byte
}

func (p patch) String() string {
	return fmt.Sprintf("(%d,%d,%s)", p.offset, p.length, p.data)
}

type Patcher struct {
	patches []patch
}

func (p *Patcher) Reset() {
	p.patches = nil
}

func (p *Patcher) Delete(offset int, length int) {
	p.patches = append(p.patches, patch{
		offset: offset,
		length: length,
	})
}

func (p *Patcher) InsertString(offset int, data string) {
	p.InsertBytes(offset, []byte(data))
}

func (p *Patcher) InsertBytes(offset int, data []byte) {
	p.patches = append(p.patches, patch{
		offset: offset,
		data:   data,
	})
}

func (p *Patcher) RewriteString(offset int, length int, data string) {
	p.RewriteBytes(offset, length, []byte(data))
}

func (p *Patcher) RewriteBytes(offset int, length int, data []byte) {
	p.patches = append(p.patches, patch{
		offset: offset,
		length: length,
		data:   data,
	})
}

func (p *Patcher) PatchString(input string) (string, error) {
	output, err := p.PatchBytes([]byte(input))
	return string(output), err
}

func (p *Patcher) PatchBytes(input []byte) ([]byte, error) {
	sort.SliceStable(p.patches, func(i, j int) bool {
		return p.patches[i].offset < p.patches[j].offset
	})

	for i, patch := range p.patches {
		if patch.offset < 0 {
			return nil, fmt.Errorf("negative offset: %v", patch)
		}
		if patch.length < 0 {
			return nil, fmt.Errorf("negative length: %v", patch)
		}
		if patch.offset+patch.length > len(input) {
			return nil, fmt.Errorf("out of range: %v", patch)
		}

		if i+1 < len(p.patches) {
			next := p.patches[i+1]
			if patch.offset+patch.length > next.offset {
				return nil, fmt.Errorf("conflict: %v vs %v", patch, next)
			}
		}
	}

	var output bytes.Buffer
	var cursor int
	for _, patch := range p.patches {
		output.Write(input[cursor:patch.offset])
		output.Write(patch.data)
		cursor = patch.offset + patch.length
	}
	output.Write(input[cursor:])
	return output.Bytes(), nil
}
