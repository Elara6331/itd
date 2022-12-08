package translit

import "testing"

func TestTransliterate(t *testing.T) {
	type testCase struct {
		name     string
		input    string
		expected string
	}

	var cases = []testCase{
		{"eASCII", "œª°«»", `oeao""`},
		{"Scandinavian", "ÆæØøÅå", "AeaeOeoeAaaa"},
		{"German", "äöüÄÖÜßẞ", "aeoeueAeOeUessSS"},
		{"Hebrew", "אבגדהוזחטיכלמנסעפצקרשתףץךםן", "abgdhuzkhtyclmns'ptskrshthftschmn"},
		{"Greek", "αάβγδεέζηήθιίϊΐκλμνξοόπρσςτυύϋΰφχψωώΑΆΒΓΔΕΈΖΗΉΘΙΊΪΚΛΜΝΞΟΌΠΡΣΤΥΎΫΦΧΨΩΏ", "aavgdeeziithiiiiklmnksooprsstyyyyfchpsooAABGDEEZIIThIIIKLMNKsOOPRSTYYYFChPsOO"},
		{"Russian", "Ёё", "Йoйo"},
		{"Ukranian", "ґєіїҐЄІЇ", "ghjeijiGhJeIJI"},
		{"Arabic", "ابتثجحخدذرزسشصضطظعغفقكلمنهويىﺓآئإؤأء٠١٢٣٤٥٦٧٨٩", "abtthj75dthrzssh99'66'33'fqklmnhwya2222220123456789"},
		{"Farsi", "پچژکگی\u200c؟٪؛،۱۲۳۴۵۶۷۸۹۰»«َُِّ", "pchzhkgy ?%;:1234567890<>eao"},
		{"Polish", "Łł", "Ll"},
		{"Lithuanian", "ąčęėįšųūž", "aceeisuuz"},
		{"Estonian", "äÄöõÖÕüÜ", "aAooOOuU"},
		{"Icelandic", "ÞþÐð", "ThthDd"},
		{"Czech", "řěýáíéóúůďťň", "reyaieouudtn"},
		{"French", "àâéèêëùüÿç", "aaeeeeuuyc"},
		{"Romanian", "ăĂâÂîÎșȘțȚşŞţŢ„”", `aAaAiIsStTsStT""`},
		{
			"Emoji",
			"😂🤣😊☺️😌😃😁😋😛😜🙃😎😶😩😕😏💜💖💗❤️💕💞💘💓💚💙💟❣️💔😱😮😯😝🤔😔😍😘😚😙👍👌🤞✌️🌄🌞🤗🌻🥱🙄🔫🥔😬✨🌌💀😅😢💯🔥😉😴💤",
			`XDXD:):):):D:D:P:P;P(:8):#-_-:(:‑J<3<3<3<3<3<3<3<3<3<3<3<3!</3D::O:OxP',:-|:|:*:*:*:*:thumbsup::ok_hand::crossed_fingers::victory_hand::sunrise_over_mountains::sun_with_face::hugging_face::sunflower::yawning_face::face_with_rolling_eyes::gun::potato::E******8-X':D:'(:100::fire:;):zzz::zzz:`,
		},
		{"Korean", "\ucc2c\ubbf8\ub97c \uc637\uc744 \uc5bc\ub9c8\ub098 \ud48d\ubd80\ud558\uac8c \uccad\ucd98\uc774 \uc5ed\uc0ac\ub97c", "chanmireul oteul eolmana pungbuhage cheongchuni yeoksareul"},
		{"Chinese", "\u81e8\u8cc7\u601d\u7531\u554f\u805e\u907f\u6c5a\u81f3\u5c0e\u524d\u99ac\u59cb\u4e00\u79fb\u3002", "lin zi si you wen wen bi wu zhi dao qian ma shi yi yi"},
		{"Armenian", "\u0531\u0532\u0533\u0534\u0535\u0536\u0537\u0538\u0539\u053a\u053b\u053c\u053d\u053e\u053f\u0540\u0541\u0542\u0543\u0544\u0545\u0546\u0547\u0548\u0549\u054a\u054b\u054c\u054d\u054e\u054f\u0550\u0551\u0552\u0553\u0554\u0555\u0556\u0561\u0562\u0563\u0564\u0565\u0566\u0567\u0568\u0569\u056a\u056b\u056c\u056d\u056e\u056f\u0570\u0571\u0572\u0573\u0574\u0575\u0576\u0577\u0578\u0579\u057a\u057b\u057c\u057d\u057e\u057f\u0580\u0581\u0582\u0583\u0584\u0585\u0586\u0587", "ABGDEZEYTJILXCKHDzXCMYNShVoChPJRSVTRCPQOFabgdezeytjilxckhdzxcmynsochpjrsvtrcpqofev"},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			out := Transliterate(tCase.input, tCase.name)
			if out != tCase.expected {
				t.Errorf("Expected %q, got %q", tCase.expected, out)
			}
		})
	}
}
