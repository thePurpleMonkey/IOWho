"use strict";

import { add_alert, getUrlParameter, alert_ajax_failure, get_session_alert, dates } from "./utilities.js";

let add_transaction = false;
let add_contact = false;

let user_id;

// Object for storing collection settings
let settingsKey = `user_${user_id}`;
let settings = undefined;


// Show options in navbar

// Enable tooltips

// Get collection info when document becomes ready
$(function() {
    // Load dashboard settings
    reloadSettings();

	// Check for any alerts
	let alert = get_session_alert();
	if (alert) {
		add_alert(alert.title, alert.message, alert.style);
    }
    
    // Show tutorial when dashboard has fully loaded
    $.when(reloadTransactions(), reloadContacts()).then(function() {
        // Dashboard fully loaded
    });
});

function name_compare(item1, item2) {
    return item1.name.localeCompare(item2.name);
}

function added_compare(item1, item2) {
    return dates.compare(item1.date_added, item2.date_added)
}

function reloadTransactions() {
    $("#transactions").empty();
    $("#transactions").append('<a class="list-group-item">Loading transactions, please wait...</a>');
    let payload = { recent: 10 };
    let result = $.Deferred();

    $.get(`/transactions`, payload)
    .done(function(data) {
        console.log("Get transactions result:");
        console.log(data);
        $("#transactions").empty();

        data.forEach(transaction => {
            let a = $("<a>");
            a.addClass("list-group-item");
            a.addClass("list-group-item-action");
            a.attr("href", `transactions.html?transaction_id=${encodeURIComponent(transaction.transaction_id)}`);
            a.text(transaction.transaction);
            $("#transactions").append(a);
        });

        result.resolve(data);
    })
    .fail(function(data) {
        if (data.status === 403) {
            window.location.replace("NotFound.html");
        }

        alert_ajax_failure("Unable to get transactions!", data);
        result.reject(data);
    })
    .always(function() {
        $("#loading").remove();
    });

    return result.promise();
};

function reloadContacts() {
    $("#contacts").empty();
    $("#hide_contact_list").empty();

    $("#contacts").append('<a class="list-group-item">Loading contacts, please wait...</a>');

    let result = $.Deferred();
    
    $.get(`/contacts`)
    .done(function(data) {
        console.log("Get contacts result:");
        console.log(data);
        $("#contacts").empty();

        data.forEach(contact => {
            let a = $("<a>");
            a.addClass("list-group-item");
            a.addClass("list-group-item-action");
            a.attr("href", `contact.html?contact_id=${encodeURIComponent(contact.contact_id)}`);
            a.text(contact.name);
            $("#contacts").append(a);
        });

        result.resolve(data);
    })
    .fail(function(data) {
        alert_ajax_failure("Unable to get contacts!", data);
        result.reject(data);
    })
    .always(function() {
        $("#loading").remove();
    });

    return result.promise();
};

function reloadSettings() {
    // Get settings from local storage
    console.log("Loading settings from key: " + settingsKey);

    let settings_string = localStorage.getItem(settingsKey);
    if (settings_string === null) {
        console.log("No settings found. Initializing settings object");
        settings = {

        };
        saveSettings();
    } else {
        try {
            console.log("Settings string: " + settings_string);
            settings = JSON.parse(settings_string);
        } catch (err) {
            console.error("Unable to load settings!");
            console.error(err);
            add_alert("Unable to load settings", "There was a problem loading the settings. Please refresh the page and try again.", "warning");
            return
        }
    }
    
    console.log("Settings:");
    console.log(settings);
}

function saveSettings() {
    let settings_string = JSON.stringify(settings);
    try {
        console.log("Saving settings to: " + settingsKey);
        localStorage.setItem(settingsKey, settings_string);
        console.log("Saved settings:");
        console.log(settings);
    } catch (err) {
        console.error("Unable to save settings to local storage!");
        console.error(err);
        add_alert("Unable to save settings", "The settings were unable to be saved. Please refresh the page and try again.", "danger");
    }
}

// #region Add transaction
function add_transaction_submit() {
    if ($("#add_transaction_form").valid()) {
        add_transaction = true;
        $("#add_transaction_modal").modal("hide");
    } else {
        console.log("Add transaction form not valid.");
    }
}
$("#add_transaction_modal_button").click(function() {
    add_transaction_submit();
});
$("#add_transaction_form").submit(function() {
    console.log("Song form submitted");
    add_transaction_submit();
    return false;
});
$('#add_transaction_modal').on('hidden.bs.modal', function (e) {
    if (add_transaction) {
        $("#transaction_wait").modal("show");
    }
});

// Make Transaction POST API call after wait dialog is shown
$('#add_transaction_wait_modal').on('shown.bs.modal', function (e) {
    let payload = JSON.stringify({
        description: $("#transaction_description").val(),
        amount: $("#transaction_amount").val(),
        transaction_timestamp: undefined, // Will be set below
        notes: $("#transaction_notes").val(),
    });

    let transaction_timestamp = $("#transaction_timestamp").val();
    if (transaction_timestamp !== "") {
        payload.transaction_timestamp = new Date(transaction_timestamp).toISOString();
    }

    $.post(`/transactions`, payload)
    .done(function(data) {
        console.log("Successfully added transaction! API response:");
        console.log(data);

        // Clear form fields
        $("#transaction_description").val("");
        $("#transaction_amount").val("");
        $("#transaction_timestamp").val("");
        $("#transaction_notes").val("");
    })
    .fail(function(data) {
        alert_ajax_failure("Unable to add transaction!", data);
        $("#add_transaction_wait_modal").modal("hide");
    })
    .always(function() {
        add_transaction = false;
        reloadSongs();
    });
});
//#endregion

// #region Add Contact
$("#add_contact_modal_button").click(function() {
    add_contact_submit();
});
$("#add_contact_form").submit(function() {
    add_contact_submit();
    return false;
});
function add_contact_submit() {
    if ($("#add_contact_form").valid()) {
        add_contact = true
        $("#add_contact_modal").modal("hide");
    }
}

$("#add_contact_modal").on("shown.bs.modal", function() {
    // Focus on name input so user can begin typing immediately
    $("#contact_name").focus();
});

$('#add_contact_modal').on('hidden.bs.modal', function (e) {
    if (add_contact) {
        add_contact = false;
        $("#add_contact_wait_modal").modal("show");
    }
});
// Make contact POST API call after contact wait dialog is shown
$('#add_contact_wait_modal').on('shown.bs.modal', function (e) {
    let payload = JSON.stringify({
        name: $("#contact_name").val(),
        email: $("#contact_email").val(),
        phone: $("#contact_phone").val(),
        notes: $("#contact_notes").val(),
    });
    console.log("Adding contact: " + payload);
    $.post(`/contacts`, payload)
    .done(function(data) {
        console.log("Add contact response:");
        console.log(data);

        // Clear add contact fields
        $("#contact_name").val("");
        $("#contact_email").val("");
        $("#contact_phone").val("");
        $("#contact_notes").val("");

        // Display success message
        add_alert("Contact created!", `${contact_name} was successfully created.`, "success", {replace_existing: true});
    })
    .fail(function(data) {
        alert_ajax_failure(`Unable to add ${contact_name}!`, data);
    })
    .always(function() {
        $("#add_contact_wait_modal").modal("hide");
        reloadContacts();
    });
});
// #endregion

// Attach to navbar buttons
$("#edit_button").click(function() {
    $("#edit_collection_modal").modal("show");
});
$("#delete_button").click(function() {
    $("#delete_collection_modal").modal("show");
});

// #region Save Settings
$("#settings_save_button").click(function() {
    settings.hidden_contacts = [];
    $("#hide_contact_list .btn-dark").each(function() {
        settings.hidden_contacts.push($(this).data("contact_id"));
    });

    // Save transaction sort order
    settings.transaction_sort = $('input[name=transaction_sort_order]:checked').val();
    settings.contact_sort = $('input[name=contact_sort_order]:checked').val();
    saveSettings();

    reloadSongs();
    reloadTags();

    $("#settings_modal").modal("hide");
});

$("#settings_modal").on("hidden.bs.modal", function() {
    reloadSettings();
});
// #endregion
