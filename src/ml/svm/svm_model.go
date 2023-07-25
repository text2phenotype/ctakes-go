package svm

const (
	CSvc       = 0
	NuSvc      = 1
	OneClass   = 2
	EpsilonSvr = 3
	NuSvr      = 4

	KernelTypeLinear      = 0
	KernelTypePoly        = 1
	KernelTypeRbf         = 2
	KernelTypeSigmoid     = 3
	KernelTypePrecomputed = 4
)

type Parameter struct {
	SvmType     int       `json:"svm_type"`
	KernelType  int       `json:"kernel_type"`
	Degree      int       `json:"degree"`
	Gamma       float64   `json:"gamma"`
	Coef0       float64   `json:"coef_0"`
	CacheSize   float64   `json:"cache_size"`
	Eps         float64   `json:"eps"`
	C           float64   `json:"c"`
	NrWeight    int       `json:"nr_weight"`
	WeightLabel []int     `json:"weight_label"`
	Weight      []float64 `json:"weight"`
	Nu          float64   `json:"nu"`
	P           float64   `json:"p"`
	Shrinking   int       `json:"shrinking"`
	Probability int       `json:"probability"`
}

func (p Parameter) OneOfTypes(types ...int) bool {
	for _, tp := range types {
		if p.SvmType == tp {
			return true
		}
	}
	return false
}

type Model struct {
	Param   Parameter   `json:"param"`
	NrClass int         `json:"nr_class"`
	L       int         `json:"l"`
	SV      [][]Node    `json:"sv"`
	SvCoef  [][]float64 `json:"sv_coef"`
	Rho     []float64   `json:"rho"`
	ProbA   []float64   `json:"prob_a"`
	ProbB   []float64   `json:"prob_b"`
	Label   []int       `json:"label"`
	NSV     []int       `json:"nsv"`
}
