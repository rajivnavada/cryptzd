package web

import (
	"cryptz/crypto"
	"html/template"
	textTemplate "text/template"
)

type messagesTemplateExtensions struct {
	Session              *SessionObject
	Messages             []crypto.Message
	Users                []crypto.User
	CurrentUser          crypto.User
	FormActionName       string
	UserIdFormFieldName  string
	SubjectFormFieldName string
	MessageFormFieldName string
	WebSocketURL         string
}

func (mte messagesTemplateExtensions) SetCurrentUser(user crypto.User) *messagesTemplateExtensions {
	mte.CurrentUser = user
	return &mte
}

var baseTemplateHtml = `<!doctype html>
<html>
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>{{ .Title }}</title>
		{{ template "HeadHTML" .Extensions }}
		<link href="https://fonts.googleapis.com/css?family=Monoton" rel="stylesheet" type="text/css">
		<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" integrity="sha384-1q8mTJOASx8j1Au+a5WDVnPi2lkFfwwEAa8hDDdjZlpLegxhjVME1fgjWPGmkzs7" crossorigin="anonymous">
		<style type="text/css">
			html, body{min-height:100%;}
			#main{width:700px;margin:0 auto;}
			.dark-bg{background-color:#666;color:#efefef;}
			.monoton{font-family:'Monoton',cursive;color:#ff6666;}
			.xlarge{font-size:5em;}
			.ctxt{text-align:center;}
			.rtxt{text-align:right;}
			.btn{border-radius:0;}
			.tmargin{margin-top:2em;}
			.hidden {display:none;}
			.nav-tabs > li > a {border-radius:0;}
			textarea.form-control, input[type="text"] { border-radius: 0; }
			textarea.form-control { resize: vertical; }
		</style>
		<style type="text/css">{{ template "HeadCSS" .Extensions }}</style>
	</head>
	<body class="{{ .BodyClasses }}">
		{{ if .ShowHeader }}<h1 class="monoton ctxt tmargin xlarge">CRYPTZ</h1>{{ end }}
		<div id="main">{{ template "BodyMain" .Extensions }}</div>
		<script src="https://code.jquery.com/jquery-2.2.1.min.js" integrity="sha256-gvQgAFzTH6trSrAWoH1iPo9Xc96QxSZ3feW6kem+O00=" crossorigin="anonymous"></script>
		<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/js/bootstrap.min.js" integrity="sha384-0mSbJDEHialfmuBBQP6A4Qrprq5OVfW37PRR3j5ELqxss1yVqOtnepnHVP9aJ7xS" crossorigin="anonymous"></script>
		{{ template "BodyAfterMain" .Extensions }}
	</body>
</html>`

var baseTemplate *template.Template

var loginTemplateHtml = `
{{ define "HeadHTML" }}{{ end }}
{{ define "HeadCSS" }}{{ end }}
{{ define "BodyMain" }}
<div class="container-fluid public_key_form tmargin">
	<div class="row">
		<div class="col-xs-10 col-xs-offset-1 ctxt">
			<form action="{{ .LoginURL }}" method="POST" enctype="application/x-www-form-urlencoded" accept-charset="UTF-8">
				<div class="form-group">
					<label for="pkey">Sign in with your public key</label>
					<textarea class="form-control" rows="12" id="pkey" name="{{ .PublicKeyFormFieldName }}" autofocus></textarea>
				</div>
				<div class="form-group">
					<button class="btn btn-default" type="submit">Sign in</button>
				</div>
			</form>
		</div>
	</div>
</div>
{{ end }}
{{ define "BodyAfterMain" }}{{ end }}
`

var loginTemplate *template.Template

var activationTemplateHtml = `
{{ define "HeadHTML" }}{{ end }}
{{ define "HeadCSS" }}{{ end }}
{{ define "BodyMain" }}
<div class="container-fluid">
	<div class="row">
		<div class="col-xs-12 tmargin">
			<p>
				An email has been sent to <code>{{ .UserEmail }}</code>.
				Decrypt the email using the private key corresponding to the public key you provided us.
				Follow the steps outlined in the email to sign in.
			</p>
			<p>
				The key you shared with us has the following fingerprint. <code>{{ .KeyFingerprint }}</code>
			<p>
		</div>
	</div>
</div>
{{ end }}
{{ define "BodyAfterMain" }}{{ end }}
`

var activationTemplate *template.Template

var activationEmailTemplateText = `
Hi {{ .UserName }},

Click on the following URL to sign in. Please note that the activation token
is tied to the session that triggered this message.

{{ .ActivationURL }}

`

var activationEmailTemplate *textTemplate.Template

var messagesTemplateHtml = `
{{ define "HeadHTML" }}{{ end }}
{{ define "HeadCSS" }}
.tabpanel{min-height:5em;}
#main { width: 100%; margin: 0; }
.upper { text-transform: uppercase; }
.left-sidebar {
	display: flex;
	flex-direction: column;
	background-color: #666;
	position: fixed;
	top: 0;
	left: 0;
	width: 132px;
	bottom: 0;
}
.left-sidebar > h3 { color: #ff6666; padding-bottom: 1em; margin-bottom: 1em; border-bottom: 1px solid #ff6666; }
.left-sidebar .link {
	padding: 0.5em 12px;
	cursor: pointer;
}
.left-sidebar .footer {
	position: absolute;
	bottom: 0;
	right: 0;
	left: 0;
	height: 50px;
	line-height: 50px;
	border-top: 1px solid #ff6666;
	color: #ff6666;
}
.left-sidebar .link:hover {
	background-color: #999;
}
.left-sidebar .link:active,
.left-sidebar .link.active {
	background-color: #fff;
}
.left-sidebar .link.active a,
.left-sidebar .link:active a {
	color: #666;
}
.left-sidebar .link:hover a {
	color: #222;
}
.left-sidebar .link a {
	color: #fff;
}

a,
a:hover,
a:visited,
a:active,
a:focus {
	text-decoration: none;
	border: 0 none;
}

.main-content { display: flex; margin-left: 132px; }
.main-content > .row { width: 100%; }
.main-content .link-content { display: none; }
.main-content .link-content.active { display: block; }
.main-content .link-content .media { cursor: pointer; }
.main-content .link-content .media-left .thumbnail { width: 64px; }
.main-content .link-content .media-heading { padding-top: 3px; }
.main-content .link-content .media-body .email { color: #888; margin-bottom: 5px; }
.main-content .link-content .form { padding: 20px; border: 1px solid #ccc; margin-top: 1em; }

.media-left .thumbnail { margin-bottom: 0; }

.link-content > div { padding: 1em 0; }

.alert.alert-danger,
.alert.alert-success {border-radius:0;margin-bottom: 6px;padding-top:0.3em;padding-bottom:0.3em;font-size:0.85em;}
{{ end }}
{{ define "BodyMain" }}
<div class="left-sidebar">
	<h3 class="ctxt upper monoton">Cryptz</h3>
	<div class="links">
		<div class="link active">
			<a href="#messages" title="Messages"><i class="glyphicon glyphicon-user"></i> Messages</a>
		</div>
		<div class="link">
			<a href="#users" title="Users"><i class="glyphicon glyphicon-envelope"></i> Users</a>
		</div>
	</div>
	<div class="footer ctxt">
		&copy; 2016
	</div>
</div>
<div class="main-content container-fluid tmargin">
	<div class="row">
		<div class="col-xs-12 col-md-6">
			<div class="link-content active" id="messages">
				{{ if .Messages }}
					{{ range $index, $message := .Messages }}
						{{ template "Message" $message }}
					{{ end }}
				{{ else }}
					<h3 class="no-messages-header">No messages for you!</h3>
				{{ end }}
			</div>
			<div class="link-content" id="users">
				{{ range $index, $user := .Users }}
					{{ with $currentUser := $.SetCurrentUser $user }}
						{{ template "User" $currentUser }}
					{{ end }}
				{{ end }}
			</div>
		</div>
	</div>
</div>
{{ end }}
{{ define "BodyAfterMain" }}
<script type="text/javascript">
$(function () {
	"use strict";

	var $links = $('.left-sidebar .links .link');
	var $linkContents = $('.main-content .link-content');
	var $users = $('#users');
	var $messages = $('#messages');

	function wireUserMedia($elements) {

		var $messageForms = $elements.siblings('.form');

		$elements.click(function (e) {
			e.preventDefault();

			var $this = $(this).siblings('.form');
			if (!$this.hasClass('hidden')) {
				$this.addClass('hidden');
				return false;
			}

			$messageForms.addClass('hidden');
			$this.
				removeClass('hidden').
				find('form').
					removeClass('disabled').
				end().
				find('button').
					removeClass('hidden disabled').
				end().
				find('.alert').
					removeClass('alert-success alert-danger').
					addClass('hidden').
				end().
				find('[type="text"], textarea').
					val('').
				end().
				find('[type="text"]').
					focus();

			return false;
		});

		$messageForms.
			find('form').
				submit(function (e) {
					var $this = $(this);
					if ($this.hasClass('disabled')) {
						return false;
					}

					$this.addClass('disabled').find('[type="submit"]').addClass('disabled');

					var action = $.trim($this.attr('action'));
					var data = $this.serialize();

					$.post(action, $this.serialize(), function (data) {
						var o = $.parseJSON(data);
						if (o.errors && o.errors.length > 0) {
							console.error(o.errors);
							$this.find('.alert').html("There were some errors. Check console for details.").removeClass("hidden").addClass('alert-danger');
						} else {
							$this.find('.alert').html("Message sent successfully").addClass('alert-success').removeClass('hidden');
						}
					}).fail(function () {
						$this.find('.alert').html("There were some errors. Check console for details.").removeClass("hidden").addClass('alert-danger error');
					});

					return false;
		});
	}

	$links.click(function (e) {
		var $this = $(this);

		e.preventDefault();
		
		// Adds / Removes active class on the link
		$links.removeClass('active');
		$this.addClass('active');

		// We need to find the link-content, this link points to and activate that.
		var href = $.trim($this.find('a').attr('href'));
		$linkContents.removeClass('active');
		$(href).addClass('active');

		return false;
	});

	wireUserMedia($users.find('.media'));

	var DEFAULT_RETRY_TIMEOUT = 1500;
	var retryTimeout = DEFAULT_RETRY_TIMEOUT;
	var wsConn;

	function doWS() {

		wsConn = new WebSocket("{{ .WebSocketURL }}");

		wsConn.onopen = function (e) {
			console.log("Connected and ready to receive");
			// Reset the retryTimeout to original value.
			retryTimeout = DEFAULT_RETRY_TIMEOUT;
		};

		wsConn.onclose = function (e) {
			console.log("Connection closed!");
			if (!retryTimeout) return;
			console.log("Will reconnect in " + retryTimeout + "ms");
			setTimeout(doWS, retryTimeout);
			retryTimeout = 2 * retryTimeout;
		};

		wsConn.onmessage = function (e) {
			var $data = $(e.data);
			var curId = $data.attr('id');
			var notification;
			var subject;
			var title;

			console.log("Message received");

			if ($data.is('.message')) {
				$messages.
					find('.no-messages-header').
						remove().
					end().
					prepend($data);

				// If we can notify try to notify
				if (Notification.permission !== "denied") {
					subject = $data.find('.media-heading').text();
					title = 'New crypt from: ' + $data.find('.email').text();
					notification = new Notification(title, { body: subject });
				}

			} else if ($data.is('.user') && curId && $("#" + curId).size() === 0) {

				$users.prepend($data);
				wireUserMedia($("#" + curId).find('.media'));

			}
		};
	}

	// Start WS
	doWS();
	
	// Catch onbeforeunload
	window.onbeforeunload = function (e) {
		console.log("Closing WS connection");
		retryTimeout = 0;
		if (wsConn) {
			wsConn.close(1000);
			wsConn = null;
		}
	};

	// Ask for notification permission
	if (Notification.permission !== "denied" && Notification.permission !== "granted") {
		Notification.requestPermission(function (perm) {
			console.log("Permission: " + perm);
		});
	}
});
</script>
{{ end }}`

var messagesTemplate *template.Template

var messageTemplateHtml = `
{{ define "Message" }}
<div class="media message">
	<div class="media-left">
		<p class="thumbnail">
			<img class="media-object" src="{{ .Sender.ImageURL }}" alt="">
		</p>
	</div>
	<div class="media-body">
		<h4 class="media-heading">{{ .Subject }}</h4>
		<p class="email">{{ .Sender.Name }} ({{ .Sender.Email }})</p>
	</div>
	<pre>{{ .Text }}</pre>
</div>
{{ end }}`

var messageTemplate *template.Template

var userTemplateHtml = `
{{ define "User" }}
<div class="user" id="user-{{ .CurrentUser.Id }}">
	<div class="media">
		<div class="media-left">
			<p class="thumbnail">
				<img class="media-object" src="{{ .CurrentUser.ImageURL }}" alt="">
			</p>
		</div>
		<div class="media-body">
			<h4 class="media-heading">
				{{ .CurrentUser.Name }}
				{{ if .Session.IsCurrentUser .CurrentUser.Id }}
					&lt;-- That's you!
				{{ end }}
			</h4>
			<p class="email">{{ .CurrentUser.Email }}</p>
			<a class="send-message-link" data-userid="{{ .CurrentUser.Id }}">Send Message</a>
		</div>
	</div>
	<div class="form message-form hidden">
		<form method="POST" action="{{ .FormActionName }}" enctype="application/x-www-form-urlencoded" accept-charset="UTF-8">
			<input type="hidden" name="{{ .UserIdFormFieldName  }}" value="{{ .CurrentUser.Id }}">
			<div class="alert hidden"></div>
			<div class="form-group">
				<label for="send-message-form-subject-{{ .CurrentUser.Id }}">Subject</label>
				<input class="form-control" type="text" id="send-message-form-subject-{{ .CurrentUser.Id }}" name="subject" placeholder="Sending you a cryptz message">
			</div>
			<div class="form-group">
				<label for="send-message-form-message-{{ .CurrentUser.Id }}">Enter your message below</label>
				<textarea class="form-control" rows="5" id="send-message-form-message-{{ .CurrentUser.Id }}" name="message" placeholder="Lorem Ipsum ..."></textarea>
			</div>
			<div class="form-group rtxt">
				<button class="btn btn-default" type="submit">Send Message</button>
			</div>
		</form>
	</div>
</div>
{{ end }}`

var userTemplate *template.Template

func init() {
	var err error

	baseTemplate, err = template.New("base").Parse(baseTemplateHtml)
	if err != nil {
		panic(err)
	}

	loginTemplate, err = template.Must(baseTemplate.Clone()).Parse(loginTemplateHtml)
	if err != nil {
		panic(err)
	}

	activationTemplate, err = template.Must(baseTemplate.Clone()).Parse(activationTemplateHtml)
	if err != nil {
		panic(err)
	}

	activationEmailTemplate, err = textTemplate.New("emailMessage").Parse(activationEmailTemplateText)
	if err != nil {
		panic(err)
	}

	messagesTemplate, err = template.Must(baseTemplate.Clone()).Parse(messagesTemplateHtml)
	if err != nil {
		panic(err)
	}
	messagesTemplate, err = messagesTemplate.Parse(messageTemplateHtml)
	if err != nil {
		panic(err)
	}
	messageTemplate, err = messagesTemplate.Parse(userTemplateHtml)
	if err != nil {
		panic(err)
	}

	messageTemplate, err = template.New("message").Parse(`{{ template "Message" . }}`)
	if err != nil {
		panic(err)
	}
	messageTemplate, err = messageTemplate.Parse(messageTemplateHtml)
	if err != nil {
		panic(err)
	}

	userTemplate, err = template.New("user").Parse(`{{ template "User" . }}`)
	if err != nil {
		panic(err)
	}
	userTemplate, err = userTemplate.Parse(userTemplateHtml)
	if err != nil {
		panic(err)
	}
}
