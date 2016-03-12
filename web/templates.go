package web

import (
	"html/template"
	textTemplate "text/template"
)

var baseTemplateHtml = `<!doctype html>
<html>
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>Zecure | A platform to securely send messages to peers</title>
		{{ template "HeadHTML" . }}
		<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" integrity="sha384-1q8mTJOASx8j1Au+a5WDVnPi2lkFfwwEAa8hDDdjZlpLegxhjVME1fgjWPGmkzs7" crossorigin="anonymous">
		<style type="text/css">
			html, body{min-height:100%;}
			#main{width:700px;margin:0 auto;}
			.ctxt{text-align:center;}
			.btn{border-radius:0;}
			.tmargin{margin-top:2em;}
			.hidden {display:none;}
			.nav-tabs > li > a {border-radius:0;}
		</style>
		<style type="text/css">{{ template "HeadCSS" }}</style>
	</head>
	<body>
		<h1 class="ctxt">Welcome to Zecure</h1>
		<div id="main">{{ template "BodyMain" . }}</div>
		<script src="https://code.jquery.com/jquery-2.2.1.min.js" integrity="sha256-gvQgAFzTH6trSrAWoH1iPo9Xc96QxSZ3feW6kem+O00=" crossorigin="anonymous"></script>
		<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/js/bootstrap.min.js" integrity="sha384-0mSbJDEHialfmuBBQP6A4Qrprq5OVfW37PRR3j5ELqxss1yVqOtnepnHVP9aJ7xS" crossorigin="anonymous"></script>
		{{ template "BodyAfterMain" . }}
	</body>
</html>`

var baseTemplate *template.Template

var loginTemplateHtml = `
{{ define "HeadHTML" }}{{ end }}
{{ define "HeadCSS" }}{{ end }}
{{ define "BodyMain" }}
<div class="container-fluid public_key_form">
	<div class="row">
		<div class="col-xs-12">
			<form action="{{ .LoginURL }}" method="POST" enctype="application/x-www-form-urlencoded" accept-charset="UTF-8">
				<div class="form-group">
					<label for="pkey">Paste your public key below</label>
					<textarea class="form-control" rows="12" id="pkey" name="{{ .PublicKeyFormFieldName }}"></textarea>
				</div>
				<div class="form-group">
					<button class="btn btn-default" type="submit">Submit</button>
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
{{ define "HeadCSS" }}.tabpanel{min-height:5em;}{{ end }}
{{ define "BodyMain" }}
<div class="container-fluid tmargin">
	<div class="row">
		<div class="col-xs-12">
			<ul class="nav nav-tabs" role="tablist">
				<li role="presentation" class="active">
					<a href="#messages" aria-controls="messages" role="tab" data-toggle="tab">Messages</a>
				</li>
				<li role="presentation">
					<a href="#users" aria-controls="users" role="tab" data-toggle="tab">Users</a>
				</li>
			</ul>
			<div class="tab-content">
				<div role="tabpanel" class="tab-pane active" id="messages">
					{{ if .Messages }}
						{{ range $index, $message := .Messages }}
							<p>{{ $message.Sender.Name }}</p>
						{{ end }}
					{{ else }}
						<h3 class="ctxt">No messages for you!</h3>
					{{ end }}
				</div>
				<div role="tabpanel" class="tab-pane" id="users">
					{{ range $index, $user := .Users }}
						<p>{{ $user.Email }}</p>
					{{ end }}
				</div>
			</div>
		</div>
	</div>
</div>
{{ end }}
{{ define "BodyAfterMain" }}
<script type="text/javascript">
$(function () {
	"use strict";
	$('[role="presentation"] a').click(function (e) {
		e.preventDefault();
		$(this).tab('show');
	});
});
</script>
{{ end }}
`

var messagesTemplate *template.Template

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
}
