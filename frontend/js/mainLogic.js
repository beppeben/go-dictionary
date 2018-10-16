var myCurrentWords = [];

$('#search-text').val('');
	
jQuery(function($) { 
  // CLEARABLE INPUT
  function tog(v){return v?'addClass':'removeClass';} 
  $(document).on('input', '.clearable', function(){
	$(this)[tog(this.value)]('x');
  }).on('mousemove', '.x', function( e ){
	$(this)[tog(this.offsetWidth-22 < e.clientX-this.getBoundingClientRect().left)]('onX');   
  }).on('touchstart click', '.onX', function( ev ){
	ev.preventDefault();
	$(this).removeClass('x onX').val('').change().keyup();
  });
});

$("#notfoundSend").click(function() {
	var url = "/services/notify?word=" + $('#search-text').val() + "&langkey=" + $('#select').val();
	$.get(url);
	$("#thanks").show();
	$('#notfoundText').hide();
});

$('#search-text').keyup( function(e) {
	if (this.value.length <= 1){
		$('#notfoundText').hide();
		$("#thanks").hide();
		$("#container").show();
	}
	$('#notfoundWord').text(this.value);
});

$('#select').change( function () {
	$('#search-text').val('');
	$('#search-text').keyup();
	$('#notfoundDictionary').text($("#select option:selected").text());
});
var myRe = new RegExp("search/(.+)/", "g");
var myArray = myRe.exec(window.location.href);
if (myArray != null) {
	$('#select').val(myArray[1]);
}

$('#mainlang').change( function () {
	window.location.replace("?lang=" + this.value);
});

$('#notfoundDictionary').text($("#select option:selected").text());

$(function() {
	$('#search-text').autoComplete({
      	minChars: 1,
		source: function(term, response){
    			$.getJSON('/services/autocomplete/' + $('#select').val(), { term: term }, function(data){ 
				//myCurrentWords = data;
				response(data);
				if (data.length > 0){
					$('#notfoundText').hide();
					myCurrentWords = data;
				} else{
					$('#notfoundText').show();
					$('#container').hide();
					$("#thanks").hide();
				}
			});
		},
		renderItem: function (item, search){
       		search = search.replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&');
		var highlighted = highlight(item.w, search);
           	return '<div class="autocomplete-suggestion" data-term="'+item.w+'" data-val="'+search+'">'+ '(' + item.t.substring(0, 2) + ') '+highlighted+'</div>';
        },
		onSelect: function(e, term, item){
           	$('#search-text').val(item.data('term'));
			searchWord(item.data('term'));
      	},
		cache: false
 	});
})
