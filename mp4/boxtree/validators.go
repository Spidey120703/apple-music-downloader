package boxtree

import "github.com/Spidey120703/go-mp4"

var BoxTypeRoot = mp4.StrToBoxType("root")

var MandatoryBoxes = map[mp4.BoxType][]mp4.BoxType{
	BoxTypeRoot: {
		mp4.BoxTypeFtyp(),
		mp4.BoxTypeMoov(),
	},
	mp4.BoxTypeMoov(): {
		mp4.BoxTypeMvhd(),
		mp4.BoxTypeTrak(),
	},
	mp4.BoxTypeTrak(): {
		mp4.BoxTypeTkhd(),
		mp4.BoxTypeMdia(),
	},
	mp4.BoxTypeMdia(): {
		mp4.BoxTypeMdhd(),
		mp4.BoxTypeHdlr(),
		mp4.BoxTypeMinf(),
	},
	mp4.BoxTypeMinf(): {
		mp4.BoxTypeDinf(),
		mp4.BoxTypeStbl(),
	},
	mp4.BoxTypeDinf(): {
		mp4.BoxTypeDref(),
	},
	mp4.BoxTypeStbl(): {
		mp4.BoxTypeStsd(),
		mp4.BoxTypeStts(),
		mp4.BoxTypeStsc(),
		mp4.BoxTypeStco(),
	},
	mp4.BoxTypeMvex(): {
		mp4.BoxTypeTrex(),
	},
	mp4.BoxTypeMoof(): {
		mp4.BoxTypeMfhd(),
	},
	mp4.BoxTypeTraf(): {
		mp4.BoxTypeTfhd(),
	},
	mp4.BoxTypeMfra(): {
		mp4.BoxTypeMfro(),
	},
	mp4.BoxTypeMeta(): {
		mp4.BoxTypeHdlr(),
	},
}

func ValidateISOBMFF(root *BoxNode) bool {
	var handler = func(node *BoxNode) bool {
		var boxType = mp4.BoxTypeAny()
		if node == root {
			boxType = BoxTypeRoot
		}
		if node.Info != nil {
			boxType = node.Info.Type
		}
		for _, mandatory := range MandatoryBoxes[boxType] {
			children, found := node.Cache[mandatory]
			if !found || len(children) == 0 {
				return false
			}
		}
		return true
	}
	var recursive func(node *BoxNode) bool
	recursive = func(node *BoxNode) bool {
		if !handler(node) {
			return false
		}
		for _, child := range node.Children {
			if !recursive(child) {
				return false
			}
		}
		return true
	}
	return recursive(root)
}
