// +build cuda

package nnops

import (
	"fmt"
	"hash"

	"github.com/chewxy/hm"
	"gorgonia.org/cu/dnn"
	"gorgonia.org/cu/dnn/interop"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

var (
	_ G.Op       = &maxpool{}
	_ G.CUDADoer = &maxpool{}
	_ G.Op       = &maxpoolDiff{}
	_ G.CUDADoer = &maxpoolDiff{}
)

type maxpool struct {
	*cudnn.Pooling

	xDesc *cudnn.TensorDescriptor
	yDesc *cudnn.TensorDescriptor
}

func newMaxPoolOp(x *gorgonia.Node, kernel, pad, stride []int) {
	var xDesc *cudnn.TensorDescriptor
	if xDesc, err = t2cudnn.Describe(x); err != nil {
		return nil, err
	}

	var p *cudnn.Pooling
	if p, err = cudnn.NewPooling(cudnn.MaxPooling, cudnn.NotPropagateNan, kernel, stride, pad); err != nil {
		return nil, err
	}
	return &maxpool{
		Pooling: p,
	}
}

func (p *maxpool) Arity() int { return 1 }

func (p *maxpool) Type() hm.Type { return hm.NewFnType(hm.TypeVariable('a'), hm.TypeVariable('a')) }

func (p *maxpool) InferShape(inputs ...gorgonia.DimSizer) (tensor.Shape, error) {
	if err = checkArity(p, len(inputs)); err != nil {
		return
	}
	return p.OutputShape(p.xDesc, 2) // only maxpool2d for now
}

func (p *maxpool) Do(...gorgonia.Value) (gorgonia.Value, error) {
	panic("not implemented")
}

func (p *maxpool) ReturnsPtr() bool { return true }

func (p *maxpool) CallsExtern() bool { return true }

func (p *maxpool) OverwritesInput() int { return -1 }

func (p *maxpool) WriteHash(h hash.Hash) {
	xShape := p.xDesc.Shape()
	kernel := p.Shape()
	padding := p.Padding()
	strides := p.Strides()
	fmt.Fprintf(h, "MaxPool{%d, %d, %d, %d}(kernel: (%d, %d), pad: (%d, %d), stride: (%d, %d))",
		xShape[0], xShape[1], xShape[2], xShape[3],
		kernel[0], kernel[1],
		padding[0], padding[1],
		strides[0], strides[1])
}

func (p *maxpool) Hashcode() uint32 { return simpleHash(p) }

func (p *maxpool) String() string {
	xShape := p.xDesc.Shape()
	kernel := p.Shape()
	padding := p.Padding()
	strides := p.Strides()
	return fmt.Sprintf("MaxPool{%d, %d, %d, %d}(kernel: (%d, %d), pad: (%d, %d), stride: (%d, %d))",
		xShape[0], xShape[1], xShape[2], xShape[3],
		kernel[0], kernel[1],
		padding[0], padding[1],
		strides[0], strides[1])
}

func (p *maxpool) CUDADo(extern gorgonia.External, dev gorgonia.Device, prealloc gorgonia.Value, inputs ...gorgonia.Value) (retVal gorgonia.Value, err error) {
	if err = checkArity(p, len(inputs)); err != nil {
		return
	}
	in := inputs[0]

	if p.yDesc == nil {
		if p.yDesc, err = t2cudnn.Describe(im); err != nil {
			return
		}
	}

	machine := extern.(G.CUDAMachine)
	ctx := machine.CUDNNContext()
	err = ctx.PoolingForward(p.Pooling, 1.0, p.xDesc, in.(cudnn.Memory), 0, p.yDesc, prealloc.(cudnn.Memory))
	return prealloc, err
}

func (p *maxpool) DiffWRT(inputs int) []bool { return []bool{true} }

func (p *maxpool) SymDiff(inputs gorgonia.Nodes, output *gorgonia.Node, grad *gorgonia.Node) (retVal gorgonia.Nodes, err error) {
	if err = checkArity(p, len(inputs)); err != nil {
		return
	}
	diff := &maxpoolDiff{p.Pooling, p.xDesc}
	x := inputs[0]

	retVal = make(G.Nodes, 1)
	retVal[0], err = G.ApplyOp(diff, x, output, grad)
	return
}

func (p *maxpool) DoDiff(ctx gorgonia.ExecutionContext, inputs gorgonia.Nodes, output *gorgonia.Node) error {
	panic("not implemented")
}

type maxpoolDiff struct {
	*cudnn.Pooling
	xDesc *cudnn.TensorDescriptor
}

func (p *maxpoolDiff) Arity() int { return 3 }

func (p *maxpoolDiff) Type() hm.Type {
	return hm.NewFnType(hm.TypeVariable('a'), hm.TypeVariable('a'), hm.TypeVariable('a'), hm.TypeVariable('a'))
}

func (p *maxpoolDiff) InferShape(inputs ...gorgonia.DimSizer) (tensor.Shape, error) {
	return inputs[0].(tensor.Shape).Clone(), nil
}

func (p *maxpoolDiff) Do(...gorgonia.Value) (gorgonia.Value, error) {
	panic("not implemented")
}

func (p *maxpoolDiff) ReturnsPtr() bool { return true }

func (p *maxpoolDiff) CallsExtern() bool { return true }

func (p *maxpoolDiff) OverwritesInput() int { return -1 }

func (p *maxpoolDiff) WriteHash(h hash.Hash) {
	xShape := p.xDesc.Shape()
	kernel := p.Shape()
	padding := p.Padding()
	strides := p.Strides()
	fmt.Fprintf(h, "MaxPoolDiff{%d, %d, %d, %d}(kernel: (%d, %d), pad: (%d, %d), stride: (%d, %d))",
		xShape[0], xShape[1], xShape[2], xShape[3],
		kernel[0], kernel[1],
		padding[0], padding[1],
		strides[0], strides[1])
}

func (p *maxpoolDiff) Hashcode() uint32 { return simpleHash(p) }

func (p *maxpoolDiff) String() string {
	xShape := p.xDesc.Shape()
	kernel := p.Shape()
	padding := p.Padding()
	strides := p.Strides()
	return fmt.Sprintf("MaxPoolDiff{%d, %d, %d, %d}(kernel: (%d, %d), pad: (%d, %d), stride: (%d, %d))",
		xShape[0], xShape[1], xShape[2], xShape[3],
		kernel[0], kernel[1],
		padding[0], padding[1],
		strides[0], strides[1])
}

func (p *maxpoolDiff) CUDADo(extern gorgonia.External, dev gorgonia.Device, prealloc gorgonia.Value, inputs ...gorgonia.Value) (retVal gorgonia.Value, err error) {
	if err = checkArity(p, len(inputs)); err != nil {
		return
	}
	x, y, dy := inputs[0], inputs[1], input[2]

	machine := extern.(G.CUDAMachine)
	ctx := machine.CUDNNContext()
	err = ctx.PoolingBackward(p.Pooling, 1.0, p.yDesc, y.(cudnn.Memory), p.yDesc, dy.(cudnn.Memory), p.xDesc, x.(cudnn.Memory), 0, p.yDesc, prealloc.(cudnn.Memory))
	return prealloc, err
}