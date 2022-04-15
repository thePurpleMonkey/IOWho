"use strict";

import { alert_ajax_failure, getUrlParameter } from "./utilities.js";

let token = getUrlParameter("token");

$(function() {
	let payload = {token: token};

	// Get invitation
	$.get(`/user/verify_email`, payload)
	.done(function(data) {
		console.log("Verify GET response:");
		console.log(data);		
		$("#success").removeClass("hidden");
		window.location.href = "/verify_identity.html";
	})
	.fail(function(data) {
		console.log("Error verifying email:");
		console.log(data);
		alert_ajax_failure("Unable to verify email.", data);
		$("#failed").removeClass("hidden");
	})
	.always(function() {
		$("#loading").addClass("hidden");
	});
});
