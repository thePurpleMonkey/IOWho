"use strict";

import { add_alert, alert_ajax_failure, getUrlParameter } from "./utilities.js";


$(function() {
	// Get URL parameter
	let recipient_email = getUrlParameter("email");
	if (recipient_email) {
		$("#email").val(recipient_email);
		$("#password").focus();
	}
});

$("#login").click(function() {
	$("#wait").modal();
});
$('#wait').on('shown.bs.modal', function (e) {
	let payload = {
		email: $("#email").val(),
		password: $("#password").val(),
		remember: $("#remember_me").prop("checked"),
	};
	console.log("Payload:");
	console.log(payload);
	$.post("/user/login", JSON.stringify(payload))
		.done(function( data ) {	
			console.log("Login response data:");
			console.log(data);
			
			// Set the user to be logged in
			try {
				window.localStorage.setItem("logged_in", true);
			} catch (err) {
				console.log("Unable to set localStorage variable 'logged_in' to true.");
				console.log(err);
			}

			// Save the user_id
			try {
				window.localStorage.setItem("user_id", data.user_id);
			} catch (err) {
				console.log("Unable to set localStorage variable 'user_id'");
				console.log(err);
			}

			// Redirect to next URL
			let redirect = new URL(window.location.href).searchParams.get("redirect");
			if (redirect === null) {
				redirect = "/dashboard.html";
			}
			redirect = decodeURIComponent(redirect);
			console.log("Redirecting to: " + redirect);
			window.location.href = redirect;
		})
		.fail(function( data ) {
			if (data.status === 401) {
				add_alert("Sign in failed.", "Username or password incorrect.", "danger", {replace_existing: true});
			} else {
				alert_ajax_failure("Sign in failed.", data, true);
			}
			console.log(data)
		})
		.always(function() {
			$("#wait").modal("hide");
		});
});

$('#password').keypress(function (e) {
	if (e.which === 13) {
		$('#login').click();
		return false;
	}
});

$('#remember_me').keypress(function (e) {
	if (e.which === 13) {
		$('#login').click();
		return false;
	}
});