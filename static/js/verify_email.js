"use strict";

import { add_alert, alert_ajax_failure } from "./utilities.js";

$(function() {
	let email;
	try {
		email = window.sessionStorage.getItem("email");
	} catch (err) {
		console.log("Unable to get email address from session storage.");
		console.log(err);
	}

	if (!email) {
		email = "Error retrieving email";
	}

	$("#email").text(email);
});

$("#verify_email_button").click(function() {
	$("#wait").modal();
});
$('#wait').on('shown.bs.modal', function (e) {
	$.post("/user/verify_email")
	.done(function(data) {
		add_alert("Verification email sent", "Your account verification email has been sent. Please allow up to 15 minutes for the email to arrive in your inbox. Check your spam messages if the email is not in your inbox.");
		$("#after").removeClass("hidden");
		$("#before").addClass("hidden");
	})
	.fail(function(data) {
		alert_ajax_failure("Unable to send verification email.", data);
	})
	.always(function() {
		$("#wait").modal("hide");
	});
});