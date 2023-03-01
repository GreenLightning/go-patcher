// Package patcher provides a system to record edits (insertions, deletions,
// replacements) and apply them to a block of data while checking for conflicts.
// Please refer to the readme for detailed explanations and examples.
package patcher

import (
	"bytes"
	"fmt"
	"sort"
)

type patch struct {
	// offset is the byte offset in the original where the patch should be applied.
	offset int
	// length is the number of bytes in the original that should be replaced by the patch.
	// Not to be confused with len(data).
	length int
	// data is the new data that should be inserted.
	data []byte
}

func (p patch) String() string {
	return fmt.Sprintf("(%d,%d,%s)", p.offset, p.length, p.data)
}

// Patcher is used to record and apply the edits.
// Do not copy a non-zero Patcher.
type Patcher struct {
	patches []patch
}

// Returns true if the patcher does not contain any edits.
// Note that zero-length operations will not create an edit.
func (p *Patcher) Empty() bool {
	return len(p.patches) == 0
}

// Reset removes all edits from the Patcher.
func (p *Patcher) Reset() {
	p.patches = nil
}

// Delete remove length bytes at the given offset.
// Offset is a byte position relative to the input data to PatchString/PatchBytes.
func (p *Patcher) Delete(offset int, length int) {
	if length != 0 {
		p.patches = append(p.patches, patch{
			offset: offset,
			length: length,
		})
	}
}

// InsertString inserts the given string at the given offset.
// Offset is a byte position relative to the input data to PatchString/PatchBytes.
// Multiple inserts to the same position are inserted one after the other
// in the order of the calls to the insert functions.
func (p *Patcher) InsertString(offset int, data string) {
	p.InsertBytes(offset, []byte(data))
}

// InsertBytes inserts the given data at the given offset.
// Offset is a byte position relative to the input data to PatchString/PatchBytes.
// The contents of data must not be changed before the call to PatchString/PatchBytes.
// Multiple inserts to the same position are inserted one after the other
// in the order of the calls to the insert functions.
func (p *Patcher) InsertBytes(offset int, data []byte) {
	if len(data) != 0 {
		p.patches = append(p.patches, patch{
			offset: offset,
			data:   data,
		})
	}
}

// RewriteString replaces length bytes at the given offset with the given string.
// Offset is a byte position relative to the input data to PatchString/PatchBytes.
func (p *Patcher) RewriteString(offset int, length int, data string) {
	p.RewriteBytes(offset, length, []byte(data))
}

// RewriteBytes replaces length bytes at the given offset with the given data.
// Offset is a byte position relative to the input data to PatchString/PatchBytes.
// The contents of data must not be changed before the call to PatchString/PatchBytes.
func (p *Patcher) RewriteBytes(offset int, length int, data []byte) {
	if length != 0 || len(data) != 0 {
		p.patches = append(p.patches, patch{
			offset: offset,
			length: length,
			data:   data,
		})
	}
}

// PatchString applies the recorded edits to the given input string.
//
// The edits remain in the Patcher and can be applied to other inputs.
// Use Reset to reset the Patcher.
//
// An error is returned if any of the edits reference positions outside the
// range of input or are conflicting with each other.
//
// The error text will reference the offending edits in the form
// (<offset>,<length-to-replace>,<text-to-insert>), e.g. (10,5,) for a
// Delete(10, 5) and (5,0,foo) for a InsertString(5, "foo").
func (p *Patcher) PatchString(input string) (string, error) {
	output, err := p.PatchBytes([]byte(input))
	return string(output), err
}

// PatchBytes applies the recorded edits to the given input.
//
// The edits remain in the Patcher and can be applied to other inputs.
// Use Reset to reset the Patcher.
//
// An error is returned if any of the edits reference positions outside the
// range of input or are conflicting with each other.
//
// The error text will reference the offending edits in the form
// (<offset>,<length-to-replace>,<text-to-insert>), e.g. (10,5,) for a
// Delete(10, 5) and (5,0,foo) for a InsertString(5, "foo").
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
