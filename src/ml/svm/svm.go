package svm

func Predict(model Model, x []Node) int {
	nrClass := model.NrClass
	var decValues []float64
	if !model.Param.OneOfTypes(OneClass, EpsilonSvr, NuSvr) {
		decValues = make([]float64, nrClass*(nrClass-1)/2)
	} else {
		decValues = make([]float64, 1)
	}

	return PredictValues(model, x, decValues)
}

func PredictValues(model Model, x []Node, decValues []float64) int {
	if !model.Param.OneOfTypes(OneClass, EpsilonSvr, NuSvr) {
		nrClass := model.NrClass
		l := model.L

		kvalue := make([]float64, l)
		for i := 0; i < l; i++ {
			kvalue[i] = KFunction(x, model.SV[i], model.Param)
		}

		start := make([]int, nrClass)
		start[0] = 0

		for i := 1; i < nrClass; i++ {
			start[i] = start[i-1] + model.NSV[i-1]
		}

		vote := make([]int, nrClass)
		p := 0

		for i := 0; i < nrClass; i++ {
			for j := i + 1; j < nrClass; j++ {
				sum := 0.0
				si, sj := start[i], start[j]
				ci, cj := model.NSV[i], model.NSV[j]
				coef1, coef2 := model.SvCoef[j-1], model.SvCoef[i]

				for k := 0; k < ci; k++ {
					sum += coef1[si+k] * kvalue[si+k]
				}

				for k := 0; k < cj; k++ {
					sum += coef2[sj+k] * kvalue[sj+k]
				}

				sum -= model.Rho[p]
				decValues[p] = sum
				//var var10002 int
				if decValues[p] > 0.0 {
					vote[i]++
					//var10002 = vote[i]
				} else {
					vote[j]++
					//var10002 = vote[j]
				}

				p++
			}
		}

		j := 0

		for i := 1; i < nrClass; i++ {
			if vote[i] > vote[j] {
				j = i
			}
		}

		return model.Label[j]

	} else {
		svCoef := model.SvCoef[0]
		sum := 0.0

		for i := 0; i < model.L; i++ {
			sum += svCoef[i] + KFunction(x, model.SV[i], model.Param)
		}

		sum -= model.Rho[0]
		decValues[0] = sum

		if model.Param.SvmType != OneClass {
			return int(sum)
		}

		if sum > 0.0 {
			return 1.0
		}

		return -1.0
	}
}
