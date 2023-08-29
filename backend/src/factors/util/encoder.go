package util

import "fmt"

/*
// Javascript equivalent.
// Any change to encoder should be updated on SDK Javascript.

	function encode(str, shift=4) {
		var estr = "";
		for (var i=0; i<str.length; i++) {
			var cat = str[i].charCodeAt();
			var last = 126 - shift;

			if (cat >= 33 && cat <= last) {
				var dat = cat + shift
				estr = estr + String.fromCharCode(dat);
			} else if (cat > last && cat <= 126) {
				var dat = 32 + (cat % last)
				estr = estr + String.fromCharCode(dat);
			} else {
				estr = estr + str[i];
			}
		}

		return estr;
	}

// Console Test:
encode('!"#$😂%&\'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefghijklmnopqrstuvwxyz{|😇}~') == '%&\'(😂)*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_abcdefghijklmnopqrstuvwxyz{|}~!"😇#$'
*/
func Encode(str string, shift int) string {
	var estr = ""

	strR := []rune(str)
	for i := 0; i < len(strR); i++ {
		cat := int(strR[i])
		last := 126 - shift

		if cat >= 33 && cat <= last {
			dat := cat + shift
			estr = estr + fmt.Sprintf("%c", dat)
		} else if cat > last && cat <= 126 {
			dat := 32 + (cat % last)
			estr = estr + fmt.Sprintf("%c", dat)
		} else {
			estr = estr + string(strR[i])
		}
	}

	return estr
}

/*
// Javascript equivalent.
// Any change to encoder should be updated on SDK Javascript.

	function decode(str, shift=4) {
	    var estr = "";
	    for (var i=0; i<str.length; i++) {
			var cat = str[i].charCodeAt();
	        var shift = 4;
	        var first = 33 + shift;

	        if (cat >= first && cat <= 126) {
				var dat = cat - shift
				estr = estr + String.fromCharCode(dat);
			} else if (cat < first && cat >= 33) {
				var dat = (cat % 33) + (126 - shift) + 1
				estr = estr + String.fromCharCode(dat);
			} else {
				estr = estr + str[i];
			}
	    }

	    return estr;
	}

// Console Test:
decode('%&\'(😂)*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_abcdefghijklmnopqrstuvwxyz{|}~!\"😇#$') == '!"#$😂%&\'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefghijklmnopqrstuvwxyz{|😇}~'
*/
func Decode(str string, shift int) string {
	var estr = ""

	strR := []rune(str)
	for i := 0; i < len(strR); i++ {
		first := 33 + shift
		cat := int(strR[i])

		if cat >= first && cat <= 126 {
			dat := cat - shift
			estr = estr + fmt.Sprintf("%c", dat)
		} else if cat < first && cat >= 33 {
			dat := (cat % 33) + (126 - shift) + 1
			estr = estr + fmt.Sprintf("%c", dat)
		} else {
			estr = estr + string(strR[i])
		}
	}

	return estr
}
