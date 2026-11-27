package fixture

type Box struct{}
type BoxDisabled struct{}
type BoxDefault struct{}

var (
	box         Box
	cacheBox    Box
	boxes       Box
	cacheBoxes  Box
	boxs        Box
	cacheBoxs   Box
	cachebox    Box // want: namingclarity: "cachebox" should be named "box" or "boxs" or "boxes" or suffixed with "Box" or "Boxs" or "Boxes"
	cacheboxes  Box // want: namingclarity: "cacheboxes" should be
	boxess      Box // want: namingclarity: "boxess" should be
	cacheBoxess Box // want: namingclarity: "cacheBoxess" should be
	cacheBoxES  Box // want: namingclarity: "cacheBoxES" should be
	es          Box // want: namingclarity: "es" should be
)

func pluralDisabled() {
	var box BoxDisabled
	var cacheBox BoxDisabled
	var boxes BoxDisabled      // want: namingclarity: "boxes" should be named "box" or suffixed with "Box"
	var cacheBoxes BoxDisabled // want: namingclarity: "cacheBoxes" should be
	_ = box
	_ = cacheBox
	_ = boxes
	_ = cacheBoxes
}

func pluralOmitted() {
	var box BoxDefault
	var cacheBox BoxDefault
	var boxes BoxDefault      // want: namingclarity: "boxes" should be named "box" or suffixed with "Box"
	var cacheBoxes BoxDefault // want: namingclarity: "cacheBoxes" should be
	_ = box
	_ = cacheBox
	_ = boxes
	_ = cacheBoxes
}
