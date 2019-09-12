/* CMPT 315 (Winter 2019)
    Assign2: Presentation Feedback System (Front-end)
     
    Author: Jacinda Shulman 
    Sources: Code examples by Nicholas Boers 
            - The render functions were modified from code from lab05 
            - XMLHTTPRequests were modified from source code example redux
                (Assessments Application)
*/
let ready = false;

let token = "";
let authUser = {
    firstName: "-1",
    accountID: -1
}
let questions: Array<QuestionSet>;
let presenters: Array<Presenter>;
let presentations: Array<Presentation>;
let currentForm: WrappedForm;
let answers: Array<Answer>;
let blankAnswers: Answer[];

let renderMC: doT.RenderFunction;
let renderOpen: doT.RenderFunction;
let renderTitle: doT.RenderFunction;

interface Presenter {
    accountID: number;
    firstName: string;
    lastName: string;
    title: string;
}

interface Presentation {
    presenterID: number;
    title: string;
    slotDate: string;
    slotTime: string;
}

interface QuestionSet {
    qID: number;
    qType: string;
    qNum: number;
    qText: string;
}

interface Answer {
    answerID: number;
    formID: number;
    qID: number;
    answerValue: string;
}

interface Form {
    formID: number;
    presenterID: number;
    evaluatorID: number;
}

interface WrappedForm {
    form: Form;
    answerIDs: Array<number>;
}

let submit = (answerID: string, value: string): void => {
    let req = new XMLHttpRequest();
    req.onload = (evt: Event): void => {
        if (req.status == 200) {
            console.log("answer posted successfully");
        }
        else {
            console.log("error message from server:", req.status);
        }
    }
    
    req.open("PUT", `http://localhost:8080/api/v1/answers/${answerID}`);
    req.setRequestHeader("Authorization", "Bearer " + token);
    req.send(value);
}

let getPresenter = (id: number): Presenter | null => {
    let i: number;
    let p: Presenter;
    for (i = 0; i < presenters.length; i++) {
        p = presenters[i];
        if (p.accountID == id) return p;
    }
    return null;
}
let getPresentation = (id: number): Presentation | null => {
    let i: number;
    let p: Presentation;
    for (i = 0; i < presentations.length; i++) {
        p = presentations[i];
        if (p.presenterID == id) {
            console.log("returning presentation:");
            console.log(p);
            return p;
        }
    }
    return null;
}

// renderQuestionSet uses the QuestionSet object to populate  
//  and render the dropdown list in the HTML
let renderQuestionSet = (): void => {
    // obtain the form template
    let tmpl = <HTMLElement>document.querySelector("#questions-template");
    if (!tmpl || !tmpl.textContent) {
        console.log("error selecting template");
        return;
    }
    
    // obtain the target for the questions-table
    let target = <HTMLElement>document.querySelector("#questions-table");
    if (!target) {
        console.log("error selecting #questions-table");
        return;
    }
    
    // render the template and copy the result into the DOM
    let renderQuestions = doT.template(tmpl.textContent);
    target.innerHTML = renderQuestions(questions);
}

// getQuestions sends the request to the server to get the list of
// questions, then calls the render functions to create the form questions in HTML
let getQuestions = (): void => {
    let req = new XMLHttpRequest();
    req.onload = (evt: Event): void => {
        // if error, display message
        if (req.status != 200) {
            console.log("error status from server:", req.status);
            return;
        }
        //If successful, populate array
        else if (req.status == 200) {
            questions = JSON.parse(req.responseText);
            console.log("questions received:");
            console.log(questions);
            renderQuestionSet();
        }
    }
    req.onerror = (evt: Event): void => {
        console.log("getQuestions: error connecting to server");
        return;
    }
    let authString = "Bearer " + token;
    
    req.open("GET", `http://localhost:8080/api/v1/questions`);
    req.setRequestHeader("Authorization", authString);
    req.send();
}

let checkAnswers = (): void => {
    if (answers == null) {
        console.log("answers were null");
        return;
    }
    let i = 0;
    for (i; i < answers.length; i++) {
        // obtain target answer in DOM
        let target = <HTMLInputElement>document.querySelector(`#q${answers[i].qID}`);
        if (!target) {
            console.log(`error selecting #q${answers[i].qID}`);
            return;
        }

        // Populate previous answers to the questions
        if (answers[i].qID == 11 || answers[i].qID == 12) {
            let a = <HTMLTextAreaElement>target.firstElementChild
            a.value = answers[i].answerValue;
        }
        else {
            //M/C questions
            let j = 0;
            let a = <HTMLInputElement>target.firstElementChild;
            // Iterate through buttons, select the one whose value = the answerValue
            for (j; j <= 4; j++) {
                if (answers[i].answerValue == "-1") {
                    break;
                }
                if (a.value == answers[i].answerValue) {
                    a.checked = true;
                    break
                }
                //go to next answer input field
                a = <HTMLInputElement>a.nextElementSibling;
            }
        }
    }
}

let renderAnswers = (answerIDs: Array<number>): void => {
    console.log("renderAnswers...");
    let i = 0;
    for (i; i < answerIDs.length; i++) {
        // obtain the target for the answers
        let target = <HTMLElement>document.querySelector(`#q${i + 1}`);
        if (!target) {
            console.log(`error selecting #q${i + 1}`);
            return;
        }
        
        // render the template and copy the result into the DOM
        if (questions[i].qType == "M/C") {
            target.innerHTML = renderMC(answerIDs[i]);
        }
        else if (questions[i].qType == "Open") {
            target.innerHTML = renderOpen(answerIDs[i]);
        }
    }
}

// getAnswers sends the request to the server to get the list of
// answers for a particular form, then calls the render function to update the view
let getAnswers = (data: WrappedForm): void => {
    console.log("getAnswers...");
    let f = data.form;

    let req = new XMLHttpRequest();
    req.onload = (evt: Event): void => {
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
    }

    req.onerror = (evt: Event): void => {
        console.log("getQuestions: error connecting to server");
        return;
    }
    let authString = "Bearer " + token;
    let URI = `http://localhost:8080/api/v1/forms/${f.formID}/answers`;

    req.open("GET", URI);
    req.setRequestHeader("Authorization", authString);
    req.send();

}

let renderForm = (data: WrappedForm): void => {
    let f = data.form;

    // fill in the title of the presentation
    // only create the render function if it is set to null
    if (renderTitle == null) {
        let titleTmpl = <HTMLElement>document.querySelector("#form-title-template");
        if (!titleTmpl || !titleTmpl.textContent) {
            console.log("error selecting form title template");
            return;
        }
        renderTitle = doT.template(titleTmpl.textContent);
    }
    // select the target for the title
    let titleTarg = <HTMLElement>document.querySelector("#form-header");
    if (!titleTarg) {
        console.log("error selecting #form-header");
        return;
    }
    
    let pres = getPresentation(f.presenterID);
    let l1 = <HTMLLabelElement>document.querySelector("#slot-date");
    let l2 = <HTMLLabelElement>document.querySelector("#slot-time");
    if (!l1 || !l2) {
        return;
    }
    if (pres != null && pres.slotDate != null && pres.slotTime != null) {
        l1.innerText = pres.slotDate;
        l2.innerText = pres.slotTime;
    }    
    // render the template and copy the result into the DOM
    let p = getPresenter(f.presenterID);
    titleTarg.innerHTML = renderTitle(p);

    getAnswers(data);

    // make the form visible
    let form = <HTMLElement>document.querySelector("#feedback-form");
    form.className = "active";
}

let postForm = (presenterID: number): void => {
    // Clear the message if it's already visible
    let msg = <HTMLLabelElement>document.querySelector("#dropdown-message");
    msg.innerText = "";

    // Server doesn't allow you to survey yourself
    if (authUser.accountID == presenterID) {
        msg.innerText = "*You can't submit a survey for yourself";
        return;
    }

    // Onload - create new form or populate with existing questions
    let req = new XMLHttpRequest();
    req.onload = (evt: Event): void => {
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
        renderForm(currentForm);
    }
    
    req.open("POST", `http://localhost:8080/api/v1/forms`);
    req.setRequestHeader("Authorization", "Bearer " + token);
    req.send(presenterID.toString());
}

// addDropdownListeners - when one menu is changed, the other is cleared
// and the feedback form is rendered according to the selection
function addDropdownListeners() {
    let p1 = <HTMLElement>document.querySelector("#presenters-dropdown");
    let p1Menu = <HTMLSelectElement>p1.firstElementChild;
    let p2 = <HTMLElement>document.querySelector("#presentations-dropdown");
    let p2Menu = <HTMLSelectElement>p2.firstElementChild;
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

// renderPresentations uses the array of Presentation object to populate  
// and render the dropdown list in the HTML
function renderPresentations() {
    console.log("rendering presenters...");

    // obtain the template from the HTML document
    let tmpl = <HTMLElement>document.querySelector("#presentations-template");
    if (!tmpl || !tmpl.textContent) {
        console.log("error selecting template");
        return;
    }
    // obtain the target in the DOM for the rendered HTML
    let target = <HTMLElement>document.querySelector("#presentations-dropdown");
    if (!target) {
        console.log("error selecting target");
        return;
    }

    // render the template and copy the result into the DOM
    let renderFunc = doT.template(tmpl.textContent);
    target.innerHTML = renderFunc(presentations);

    // add listener if the other menu is already loaded
    if (ready) {
        addDropdownListeners();
    } // else indicate this one is ready
    else {
        ready = true;
    }
}

// getPresentations sends the request to the server to get the list of
// presentations, then calls the render functions to create the dropdown lists in HTML
function getPresentations() {
    let req = new XMLHttpRequest();
    
    req.onload = (evt: Event): void => {
        // if error, display message
        if (req.status != 200) {
            console.log("error status from server:", req.status);
            return;
        }
        //If successful, populate array
        else if (req.status == 200) {
            presentations = JSON.parse(req.responseText);
            console.log(presentations);
            renderPresentations();
        }
    }
    
    req.onerror = (evt: Event): void => {
        console.log("getPresentations: error connecting to server");
        return;
    }
    
    let authString = "Bearer " + token;
    req.open("GET", `http://localhost:8080/api/v1/presentations`);
    req.setRequestHeader("Authorization", authString);
    req.send();
}

// renderPresenters uses the array of Presenter object to populate  
//  and render the dropdown list in the HTML
function renderPresenters() {
    // obtain the template from the HTML document
    let tmpl = <HTMLElement>document.querySelector("#presenters-template");
    if (!tmpl || !tmpl.textContent) {
        console.log("error selecting template");
        return;
    }
    
    // obtain the target in the DOM for the rendered HTML
    let target = <HTMLElement>document.querySelector("#presenters-dropdown");
    if (!target) {
        console.log("error selecting target");
        return;
    }
    
    // render the template and copy the result into the DOM
    let renderFunc = doT.template(tmpl.textContent);
    target.innerHTML = renderFunc(presenters);

    // add listener if the other menu is already loaded
    if (ready) {
        addDropdownListeners();
    } // else indicate this one is ready
    else {
        ready = true;
    }
}

// getPresenters sends the request to the server to get the list of
// presenters, then calls the render functions to create the dropdown lists in HTML
function getPresenters() {
    let req = new XMLHttpRequest();
    req.onload = (evt: Event): void => {
        // if error, display message
        if (req.status != 200) {
            console.log("error status from server:", req.status);
            return;
        }
        // If successful, populate array
        else if (req.status == 200) {
            presenters = JSON.parse(req.responseText);
            console.log(presenters);
            renderPresenters();
        }
    }  
    req.onerror = (evt: Event): void => {
        console.log("getPresenters: error connecting to server");
        return;
    }
    
    let authString = "Bearer " + token;
    req.open("GET", `http://localhost:8080/api/v1/presenters`);
    req.setRequestHeader("Authorization", authString);
    req.send();
}

let loadHomePage = (): void => {
    // hide the log in page and load the elements of the main menu
    let loginPage = <HTMLElement>document.querySelector("#login");
    let mainMenu = <HTMLElement>document.querySelector("#main-menu");
    let displayName = <HTMLLabelElement>document.querySelector("#current-user-name");
    
    displayName.innerText = authUser.firstName;
    loginPage.className = "";
    mainMenu.className = "active";

    getQuestions();
    getPresenters();
    getPresentations();
}

// authenticate sends a request to the server to check that the token is valid. 
// If successful, the response includes the id for the person signing in
let authenthicate = (inputToken: string): void => {
    token = ""; // clear the token stored

    let req = new XMLHttpRequest();
    req.onload = (evt: Event): void => {
        // If unauthorized, display message
        if (req.status == 401) {
            console.log("authenticate function got 401 back from server");
            let badTokenMessage = <HTMLLabelElement>document.querySelector("#bad-token-message");
            badTokenMessage.innerText = "*Invalid token provided";
            return;
        }
        // If authentication successful, load home page
        else if (req.status == 200) {
            authUser = JSON.parse(req.responseText);
            console.log(authUser);
            token = inputToken;
            loadHomePage();
        }
    }
    req.onerror = (evt: Event): void => {
        console.log("Error connecting to server");
        return;
    }
   
    let authString = "Bearer " + inputToken;
    req.open("GET", `http://localhost:8080/api/v1/tokens`);
    req.setRequestHeader("Authorization", authString);
    req.send();
}

let processLogIn = (): void => {
    // If input is empty, display message
    let badTokenMessage = <HTMLLabelElement>document.querySelector("#bad-token-message");
    let input = (document.querySelector("#token") as HTMLInputElement).value;
    
    if (input == null || input === "") {
        badTokenMessage.innerText = "*Please enter a valid token";
        console.log("Input field was was blank");
        return;
    }
    
    authenthicate(input);
}

window.onload = (): void => {
    // add listener to login button
    let loginBtn = <HTMLElement>document.querySelector("#login-btn");
    loginBtn.onclick = processLogIn;

    //add listenter to save button on the survey form
    let saveBtn = <HTMLButtonElement>document.querySelector("#save-btn");
    saveBtn.onclick = (evt) => {
        let form = <HTMLElement>document.querySelector("#feedback-form");
        form.className = "";
    }

    let tmpl = <HTMLElement>document.querySelector("#mc-template");
    if (!tmpl || !tmpl.textContent) {
        console.log("error selecting form title template");
        return;
    }
    renderMC = doT.template(tmpl.textContent);

    let tmpl2 = <HTMLElement>document.querySelector("#open-template");
    if (!tmpl2 || !tmpl2.textContent) {
        console.log("error selecting form title template");
        return;
    }
    renderOpen = doT.template(tmpl2.textContent);
}
