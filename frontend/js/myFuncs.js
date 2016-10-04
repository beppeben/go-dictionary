function redirectLang() {
	var baseLang = qs("lang");
	if (!baseLang) {
		var userLang = navigator.language || navigator.userLanguage;

		switch(userLang.split('-')[0]) {
			case "it": baseLang = "ita"; break;
			case "fr": baseLang = "fre"; break;
			case "de": baseLang = "ger"; break;
			case "es": baseLang = "spa"; break;
			case "ru": baseLang = "rus"; break;
			case "ja": baseLang = "jap"; break;
			case "zh": baseLang = "chi"; break;
		} 

		if (baseLang) {
			window.location.replace("?lang=" + baseLang);
		}
	}
}


function qs(key) {
	key = key.replace(/[*+?^$.\[\]{}()|\\\/]/g, "\\$&");
	var match = location.search.match(new RegExp("[?&]"+key+"=([^&]+)(&|$)"));
	return match && decodeURIComponent(match[1].replace(/\+/g, " "));
}
