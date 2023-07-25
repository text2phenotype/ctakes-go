package lemmatizer

import "regexp"

const (
	urlResult = "#url#"
	NN        = "NN"
	PRP       = "PRP"
	WP        = "WP"
	VB        = "VB"
	JJ        = "JJ"
	RB        = "RB"
	WRB       = "WRB"
)

type MTLib interface {
	normalizeBasic(form string) string
}

type mtLib struct {
	DigitLike   *regexp.Regexp
	DigitSpan   *regexp.Regexp
	UrlSpan     *regexp.Regexp
	PunctRepeat *regexp.Regexp
}

func (lib *mtLib) normalizeBasic(form string) string {
	if lib.containsURL(form) {
		return urlResult
	}

	form = lib.normalizeDigits(form)
	form = lib.normalizePunctuation(form)

	return form
}

func (lib *mtLib) containsURL(form string) bool {
	return lib.UrlSpan.MatchString(form)
}

func (lib *mtLib) normalizeDigits(form string) string {
	form = lib.DigitLike.ReplaceAllString(form, "0")
	return lib.DigitSpan.ReplaceAllString(form, "0")
}

func (lib *mtLib) normalizePunctuation(form string) string {
	return lib.PunctRepeat.ReplaceAllStringFunc(form, func(s string) string {
		return s[0:2]
	})
}

func NewMTLib() (MTLib, error) {
	var lib mtLib

	re, err := regexp.Compile(`\d%|\$\d|(^|\d)\.\d|\d,\d|\d:\d|\d-\d|\d\/\d`)
	if err != nil {
		return nil, err
	}
	lib.DigitLike = re

	re, err = regexp.Compile(`\d+`)
	if err != nil {
		return nil, err
	}
	lib.DigitSpan = re

	re, err = regexp.Compile(`((([A-Za-z]{3,9}:(?:\/\/)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~%\/.\w-_]*)?\??(?:[-\+=&;%@.\w_]*)#?(?:[.\!\/\\w]*))?|(\w+\.)+(com|edu|gov|int|mil|net|org|biz)$)`)
	if err != nil {
		return nil, err
	}
	lib.UrlSpan = re

	re, err = regexp.Compile(`\.{2,}|\!{2,}|\?{2,}|\-{2,}|\*{2,}|\={2,}|\~{2,}|\,{2,}`)
	if err != nil {
		return nil, err
	}
	lib.PunctRepeat = re

	return &lib, nil
}

func IsNoun(pos string) bool {
	return StartsWithAny(pos, NN) || AnyOf(pos, PRP, WP)
}

func IsVerb(pos string) bool {
	return StartsWithAny(pos, VB)
}

func IsAdjective(pos string) bool {
	return StartsWithAny(pos, JJ)
}

func IsAdverb(pos string) bool {
	return StartsWithAny(pos, RB) || AnyOf(pos, WRB)
}
