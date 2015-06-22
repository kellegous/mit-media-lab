(function() {

$('form').on('submit', function(e) {
	e.preventDefault();

	$('button').prop('disabled', true);
	$('#stat').addClass('load')
		.slideDown();

	$.ajax({
		url: '/api/v1/invite-me',
		dataType: 'json',
		method: 'POST',
		data: {
			email: $('#email').val(),
			first: $('#first').val(),
			last: $('#last').val()
		}
	}).done(function(data) {
		console.log(data);
		$('button').prop('disabled', false);
	});
});

})();