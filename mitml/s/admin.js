(function() {

var $e = function(name) {
	return $(document.createElement(name));
};

var nameFor = function(req) {
	if (!req.First && !req.Last) {
		return '(none)';
	} else {
		var name = req.First || '';
		return req.Last ? name + ' ' + req.Last : name;
	}
};

var formatTime = function(t) {
	var pad = function(n) {
		n += '';
		return n.length == 1 ? '0' + n : n;
	}

	return (1900 + t.getYear()) + '.'
		+ pad(t.getMonth() + 1) + '.'
		+ pad(t.getDate()) + ' '
		+ pad(t.getHours()) + ':'
		+ pad(t.getMinutes());
};

var approve = function(req, cb) {
	$.ajax({
		url: '/api/v1/approve-request',
		data: {
			key: req.Key
		},
		dataType: 'json',
		method: 'POST'
	}).done(function(data) {
		debugger;
		cb(data.ok, data.error);
	});
};

var main = function() {
	var $reqs = $('#reqs');

	$.ajax({
		url: '/api/v1/requests',
		dataType: 'json',
		method: 'GET'
	}).done(function(data) {
		data.forEach(function(req, i) {
			var $req = $e('div').addClass('req')
				.addClass((i&1) == 0 ? 'e' : 'o');

			var $b = $e('div').addClass('invite')
				.appendTo($req)
				.on('click', function(e) {
					approve(req, function(ok, error) {
						$b.addClass(ok ? 'ok' : 'no');
					});
				});
			$e('span').addClass('email')
				.text(req.Request.Email)
				.appendTo($req);
			$e('span').addClass('name')
				.text(nameFor(req.Request))
				.appendTo($req);
			$e('span').addClass('time')
				.text(formatTime(new Date(Date.parse(req.Request.Time))))
				.appendTo($req);

			if (req.Request.Invited) {
				$b.addClass(req.Request.Succeeded ? 'ok' : 'no');
			}

			$req.appendTo($reqs);
		});

	});
};

main();

})();