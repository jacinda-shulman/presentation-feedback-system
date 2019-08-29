"use strict";
var ready = false;
var token = "";
var authUser = {
    firstName: "-1",
    accountID: -1
};
var questions;
var presenters;
var presentations;
var currentForm;
var answers;
var blankAnswers;
var renderMC;
var renderOpen;
var renderTitle;
var submit = function (answerID, value) {
    var req = new XMLHttpRequest();
    req.onload = function (evt) {
        if (req.status == 200) {
            console.log("answer posted successfully");
        }
        else {
            console.log("error message from server:", req.status);
        }
    };
    req.open("PUT", "http://localhost:8080/api/v1/answers/" + answerID);
    req.setRequestHeader("Authorization", "Bearer " + token);
    req.send(value);
};
var getPresenter = function (id) {
    var i;
    var p;
    for (i = 0; i < presenters.length; i++) {
        p = presenters[i];
        if (p.accountID == id)
            return p;
    }
    return null;
};
var getPresentation = function (id) {
    var i;
    var p;
    for (i = 0; i < presentations.length; i++) {
        p = presentations[i];
        if (p.presenterID == id) {
            console.log("returning presentation:");
            console.log(p);
            return p;
        }
    }
    return null;
};
var renderQuestionSet = function () {
    console.log("rendering questions...");
    var tmpl = document.querySelector("#questions-template");
    if (!tmpl || !tmpl.textContent) {
        console.log("error selecting template");
        return;
    }
    var target = document.querySelector("#questions-table");
    if (!target) {
        console.log("error selecting #questions-table");
        return;
    }
    var renderQuestions = doT.template(tmpl.textContent);
    target.innerHTML = renderQuestions(questions);
};
var getQuestions = function () {
    console.log("getQuestions");
    var req = new XMLHttpRequest();
    req.onload = function (evt) {
        if (req.status != 200) {
            console.log("error status from server:", req.status);
            return;
        }
        else if (req.status == 200) {
            questions = JSON.parse(req.responseText);
            console.log("questions received:");
            console.log(questions);
            renderQuestionSet();
        }
    };
    req.onerror = function (evt) {
        console.log("getQuestions: error connecting to server");
        return;
    };
    var authString = "Bearer " + token;
    req.open("GET", "http://localhost:8080/api/v1/questions");
    req.setRequestHeader("Authorization", authString);
    req.send();
};
var checkAnswers = function () {
    if (answers == null) {
        console.log("answers were null");
        return;
    }
    var i = 0;
    for (i; i < answers.length; i++) {
        var target = document.querySelector("#q" + answers[i].qID);
        if (!target) {
            console.log("error selecting #q" + answers[i].qID);
            return;
        }
        if (answers[i].qID == 11 || answers[i].qID == 12) {
            var a = target.firstElementChild;
            a.value = answers[i].answerValue;
        }
        else {
            var j = 0;
            var a = target.firstElementChild;
            for (j; j <= 4; j++) {
                if (answers[i].answerValue == "-1") {
                    break;
                }
                if (a.value == answers[i].answerValue) {
                    a.checked = true;
                    break;
                }
                a = a.nextElementSibling;
            }
        }
    }
};
var renderAnswers = function (answerIDs) {
    console.log("renderAnswers...");
    var i = 0;
    for (i; i < answerIDs.length; i++) {
        var target = document.querySelector("#q" + (i + 1));
        if (!target) {
            console.log("error selecting #q" + (i + 1));
            return;
        }
        if (questions[i].qType == "M/C") {
            target.innerHTML = renderMC(answerIDs[i]);
        }
        else if (questions[i].qType == "Open") {
            target.innerHTML = renderOpen(answerIDs[i]);
        }
    }
};
var getAnswers = function (data) {
    console.log("getAnswers...");
    var f = data.form;
    var req = new XMLHttpRequest();
    req.onload = function (evt) {
        if (req.status == 200) {
            answers = JSON.parse(req.responseText);
            console.log("answers received:");
            console.log(answers);
            renderAnswers(data.answerIDs);
            checkAnswers();
        }
        else {
            renderAnswers(data.answerIDs);
        }
    };
    req.onerror = function (evt) {
        console.log("getQuestions: error connecting to server");
        return;
    };
    var authString = "Bearer " + token;
    var URI = "http://localhost:8080/api/v1/forms/" + f.formID + "/answers";
    console.log("URI:", URI);
    req.open("GET", URI);
    req.setRequestHeader("Authorization", authString);
    req.send();
};
var renderForm = function (data) {
    var f = data.form;
    console.log("rendering form with form ID:", f.formID);
    if (renderTitle == null) {
        var titleTmpl = document.querySelector("#form-title-template");
        if (!titleTmpl || !titleTmpl.textContent) {
            console.log("error selecting form title template");
            return;
        }
        renderTitle = doT.template(titleTmpl.textContent);
    }
    var titleTarg = document.querySelector("#form-header");
    if (!titleTarg) {
        console.log("error selecting #form-header");
        return;
    }
    var pres = getPresentation(f.presenterID);
    console.log(presentations);
    console.log(pres);
    var l1 = document.querySelector("#slot-date");
    var l2 = document.querySelector("#slot-time");
    if (!l1 || !l2) {
        return;
    }
    if (pres != null && pres.slotDate != null && pres.slotTime != null) {
        l1.innerText = pres.slotDate;
        l2.innerText = pres.slotTime;
    }
    var p = getPresenter(f.presenterID);
    titleTarg.innerHTML = renderTitle(p);
    getAnswers(data);
    var form = document.querySelector("#feedback-form");
    form.className = "active";
};
var postForm = function (presenterID) {
    var msg = document.querySelector("#dropdown-message");
    msg.innerText = "";
    if (authUser.accountID == presenterID) {
        msg.innerText = "*You can't submit a survey for yourself";
        return;
    }
    var req = new XMLHttpRequest();
    req.onload = function (evt) {
        if (req.status != 201 && req.status != 409) {
            console.log("error message from server:", req.status);
            return;
        }
        if (req.status == 201) {
            console.log("new form created");
        }
        else if (req.status == 409) {
            console.log("form already exists");
        }
        currentForm = JSON.parse(req.responseText);
        console.log(currentForm);
        renderForm(currentForm);
    };
    req.open("POST", "http://localhost:8080/api/v1/forms");
    req.setRequestHeader("Authorization", "Bearer " + token);
    req.send(presenterID.toString());
};
function addDropdownListeners() {
    var p1 = document.querySelector("#presenters-dropdown");
    var p1Menu = p1.firstElementChild;
    var p2 = document.querySelector("#presentations-dropdown");
    var p2Menu = p2.firstElementChild;
    p1Menu.onchange =
        function (evt) {
            p2Menu.selectedIndex = 0;
            postForm(p1Menu.selectedIndex);
        };
    p2Menu.onchange =
        function (evt) {
            p1Menu.selectedIndex = 0;
            postForm(p2Menu.selectedIndex);
        };
}
function renderPresentations() {
    console.log("rendering presenters...");
    var tmpl = document.querySelector("#presentations-template");
    if (!tmpl || !tmpl.textContent) {
        console.log("error selecting template");
        return;
    }
    var target = document.querySelector("#presentations-dropdown");
    if (!target) {
        console.log("error selecting target");
        return;
    }
    var renderFunc = doT.template(tmpl.textContent);
    target.innerHTML = renderFunc(presentations);
    if (ready) {
        addDropdownListeners();
    }
    else {
        ready = true;
    }
}
function getPresentations() {
    var req = new XMLHttpRequest();
    req.onload = function (evt) {
        if (req.status != 200) {
            console.log("error status from server:", req.status);
            return;
        }
        else if (req.status == 200) {
            presentations = JSON.parse(req.responseText);
            console.log(presentations);
            renderPresentations();
        }
    };
    req.onerror = function (evt) {
        console.log("getPresentations: error connecting to server");
        return;
    };
    var authString = "Bearer " + token;
    req.open("GET", "http://localhost:8080/api/v1/presentations");
    req.setRequestHeader("Authorization", authString);
    req.send();
}
function renderPresenters() {
    console.log("rendering presenters...");
    var tmpl = document.querySelector("#presenters-template");
    if (!tmpl || !tmpl.textContent) {
        console.log("error selecting template");
        return;
    }
    var target = document.querySelector("#presenters-dropdown");
    if (!target) {
        console.log("error selecting target");
        return;
    }
    var renderFunc = doT.template(tmpl.textContent);
    target.innerHTML = renderFunc(presenters);
    if (ready) {
        addDropdownListeners();
    }
    else {
        ready = true;
    }
}
function getPresenters() {
    var req = new XMLHttpRequest();
    req.onload = function (evt) {
        if (req.status != 200) {
            console.log("error status from server:", req.status);
            return;
        }
        else if (req.status == 200) {
            presenters = JSON.parse(req.responseText);
            console.log(presenters);
            renderPresenters();
        }
    };
    req.onerror = function (evt) {
        console.log("getPresenters: error connecting to server");
        return;
    };
    var authString = "Bearer " + token;
    req.open("GET", "http://localhost:8080/api/v1/presenters");
    req.setRequestHeader("Authorization", authString);
    req.send();
}
var loadHomePage = function () {
    var loginPage = document.querySelector("#login");
    var mainMenu = document.querySelector("#main-menu");
    var displayName = document.querySelector("#current-user-name");
    displayName.innerText = authUser.firstName;
    loginPage.className = "";
    mainMenu.className = "active";
    getQuestions();
    getPresenters();
    getPresentations();
};
var authenthicate = function (inputToken) {
    token = "";
    var req = new XMLHttpRequest();
    req.onload = function (evt) {
        if (req.status == 401) {
            console.log("authenticate function got 401 back from server");
            var badTokenMessage = document.querySelector("#bad-token-message");
            badTokenMessage.innerText = "*Invalid token provided";
            return;
        }
        else if (req.status == 200) {
            authUser = JSON.parse(req.responseText);
            console.log(authUser);
            token = inputToken;
            loadHomePage();
        }
    };
    req.onerror = function (evt) {
        console.log("Error connecting to server");
        return;
    };
    var authString = "Bearer " + inputToken;
    req.open("GET", "http://localhost:8080/api/v1/tokens");
    req.setRequestHeader("Authorization", authString);
    req.send();
};
var processLogIn = function () {
    var badTokenMessage = document.querySelector("#bad-token-message");
    var input = document.querySelector("#token").value;
    if (input == null || input === "") {
        badTokenMessage.innerText = "*Please enter a valid token";
        console.log("Input field was was blank");
        return;
    }
    authenthicate(input);
};
window.onload = function () {
    var loginBtn = document.querySelector("#login-btn");
    loginBtn.onclick = processLogIn;
    var saveBtn = document.querySelector("#save-btn");
    saveBtn.onclick = function (evt) {
        var form = document.querySelector("#feedback-form");
        form.className = "";
    };
    var tmpl = document.querySelector("#mc-template");
    if (!tmpl || !tmpl.textContent) {
        console.log("error selecting form title template");
        return;
    }
    renderMC = doT.template(tmpl.textContent);
    var tmpl2 = document.querySelector("#open-template");
    if (!tmpl2 || !tmpl2.textContent) {
        console.log("error selecting form title template");
        return;
    }
    renderOpen = doT.template(tmpl2.textContent);
};
