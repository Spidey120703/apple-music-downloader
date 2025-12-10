package cmaf

import (
	"downloader/internal/media/mp4/boxtree"
	"io"

	"github.com/Spidey120703/go-mp4"
)

type IContext interface {
	Initialize(io.ReadSeeker) error
	Finalize(io.WriteSeeker) error
	MergeSegment(io.ReadSeeker) (*Segment, error)
	GetRoot() *boxtree.BoxNode
}

type Context struct {
	Root     *boxtree.BoxNode
	Header   *Header
	Segments []*Segment
}

func (ctx *Context) Initialize(input io.ReadSeeker) (err error) {
	if ctx.Root, err = boxtree.Unmarshal(input); err != nil {
		return
	}
	if ctx.Header, err = InitializeHeader(ctx.Root); err != nil {
		return
	}
	return
}

func (ctx *Context) Finalize(output io.WriteSeeker) (err error) {
	writer := mp4.NewWriter(output)
	_, err = boxtree.Marshal(writer, ctx.Root)
	return
}

func (ctx *Context) MergeSegment(input io.ReadSeeker) (seg *Segment, err error) {
	var root *boxtree.BoxNode
	if root, err = boxtree.Unmarshal(input); err != nil {
		return
	}
	if seg, err = InitializeSegment(root); err != nil {
		return
	}
	ctx.Segments = append(ctx.Segments, seg)
	ctx.Root.Children = append(ctx.Root.Children, root.Children...)
	err = ctx.Root.Caching()
	return
}

func (ctx *Context) GetRoot() *boxtree.BoxNode {
	return ctx.Root
}
