package svm

import "math"

func KFunction(x []Node, y []Node, param Parameter) float64 {
	switch param.KernelType {
	case KernelTypeLinear:
		return dot(x, y)
	case KernelTypePoly:
		return powi(param.Gamma*dot(x, y)+param.Coef0, param.Degree)
	case KernelTypeRbf:
		{
			sum := 0.0
			xlen := len(x)
			ylen := len(y)
			i := 0
			j := 0

			for i < xlen && j < ylen {

				switch {
				case x[i].Index == y[j].Index:
					{
						i++
						j++
						d := x[i].Value - y[j].Value
						sum += d * d
					}
				case x[i].Index > y[j].Index:
					{
						sum += y[j].Value * y[j].Value
						j++
					}
				default:
					{
						sum += x[i].Value * x[i].Value
						i++
					}
				}

			}

			for i < xlen {
				sum += x[i].Value * x[i].Value
				i++
			}

			for j < ylen {
				sum += y[j].Value * y[j].Value
				j++
			}

			return math.Exp(-param.Gamma * sum)
		}
	case KernelTypeSigmoid:
		return math.Tanh(param.Gamma*dot(x, y) + param.Coef0)
	case KernelTypePrecomputed:
		return x[int(y[0].Value)].Value
	default:
		return 0.0
	}

}

func dot(x []Node, y []Node) float64 {
	sum := 0.0
	xLen := len(x)
	yLen := len(y)
	i, j := 0, 0

	for i < xLen && j < yLen {
		switch {
		case x[i].Index == y[j].Index:
			{
				sum += x[i].Value * y[j].Value
				i++
				j++
			}
		case x[i].Index > y[j].Index:
			{
				j++
			}
		default:
			{
				i++
			}
		}
	}

	return sum
}

func powi(base float64, times int) float64 {
	tmp := base
	ret := 1.0

	for t := times; t > 0; t /= 2 {
		if t%2 == 1 {
			ret *= tmp
		}

		tmp *= tmp
	}

	return ret
}
