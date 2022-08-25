// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package fb

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type IndexColumn struct {
	_tab flatbuffers.Table
}

func GetRootAsIndexColumn(buf []byte, offset flatbuffers.UOffsetT) *IndexColumn {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &IndexColumn{}
	x.Init(buf, n+offset)
	return x
}

func GetSizePrefixedRootAsIndexColumn(buf []byte, offset flatbuffers.UOffsetT) *IndexColumn {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &IndexColumn{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func (rcv *IndexColumn) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *IndexColumn) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *IndexColumn) Name(obj *CIStr) *CIStr {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(CIStr)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *IndexColumn) Offset() int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *IndexColumn) MutateOffset(n int64) bool {
	return rcv._tab.MutateInt64Slot(6, n)
}

func IndexColumnStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func IndexColumnAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(name), 0)
}
func IndexColumnAddOffset(builder *flatbuffers.Builder, offset int64) {
	builder.PrependInt64Slot(1, offset, 0)
}
func IndexColumnEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}