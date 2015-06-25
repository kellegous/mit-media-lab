(function() {

$('#form').on('submit', function(e) {
	e.preventDefault();

	var button = $('#go').prop('disabled', true),
		done = $('#done');

	$.ajax({
		url: '/api/v1/invite-me',
		method: 'POST',
		dataType: 'json',
		data: {
			email: $('#email').val(),
			first: $('#first').val(),
			last: $('#last').val()
		}
	}).done(function(data) {
		button.prop('disabled', false);
		if (data.ok) {
			done.removeClass('no')
				.addClass('ok')
				.text('You\'re all set. You should receive an invite soon.');
		} else {
			done.removeClass('ok')
				.addClass('no')
				.text(data.error);
		}
	});
});

})();