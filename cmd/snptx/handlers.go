package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tullo/snptx/internal/models"
	"github.com/tullo/snptx/internal/validator"
)

func (a *app) home(w http.ResponseWriter, r *http.Request) {
	ss, err := a.snippets.Latest(r.Context())
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	data := a.newTemplateData(r)
	data.Snippets = ss

	a.render(w, r, http.StatusOK, "home.tmpl", data)
}

func (a *app) about(w http.ResponseWriter, r *http.Request) {
	data := a.newTemplateData(r)
	a.render(w, r, http.StatusOK, "about.tmpl", data)
}

func (a *app) snippetDeletePost(w http.ResponseWriter, r *http.Request) {
	//id := r.URL.Query().Get(":id")
	id := r.PathValue("id")
	err := a.snippets.Delete(r.Context(), id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.sessionManager.Put(r.Context(), "flash", "Snippet successfully deleted!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *app) snippetView(w http.ResponseWriter, r *http.Request) {
	// pat does not strip the colon from the named capture key,
	// get the value of ":id" from the query string instead of "id"
	//id := r.URL.Query().Get(":id")
	id := r.PathValue("id")
	s, err := a.snippets.Retrieve(r.Context(), id)
	if err != nil {
		// unwrapping errors
		if errors.Is(err, models.ErrNoRecord) {
			a.notFound(w)
		} else {
			a.serverError(w, r, err)
		}
		return
	}

	data := a.newTemplateData(r)
	data.Snippet = s

	a.render(w, r, http.StatusOK, "view.tmpl", data)
}

type snippetEditForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	validator.Validator `form:"-"`
}

func (a *app) updateSnippetForm(w http.ResponseWriter, r *http.Request) {
	//id := r.URL.Query().Get(":id")
	id := r.PathValue("id")
	s, err := a.snippets.Retrieve(r.Context(), id)
	if err != nil {
		// unwrapping errors
		if errors.Is(err, models.ErrNoRecord) {
			a.notFound(w)
		} else {
			a.serverError(w, r, err)
		}
		return
	}

	data := a.newTemplateData(r)
	data.Snippet = s
	data.Form = snippetEditForm{
		Title:   s.Title,
		Content: s.Content,
	}

	a.render(w, r, http.StatusOK, "edit.tmpl", data)
}

func (a *app) updateSnippetPost(w http.ResponseWriter, r *http.Request) {

	var form snippetEditForm

	err := a.decodePostForm(r, &form)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")

	if !form.Valid() {
		data := a.newTemplateData(r)
		data.Form = form
		a.render(w, r, http.StatusUnprocessableEntity, "edit.tmpl", data)
		return
	}

	// update snippet record in the database using the form data
	up := models.UpdateSnippet{
		Title:   &form.Title,
		Content: &form.Content,
	}

	id := r.PathValue("id")
	err = a.snippets.Update(r.Context(), id, up, time.Now().Local())
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.sessionManager.Put(r.Context(), "flash", "Snippet successfully updated!")
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%s", id), http.StatusSeeOther)
}

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

func (a *app) snippetCreateForm(w http.ResponseWriter, r *http.Request) {
	// render the form using an empty forms.Form struct
	data := a.newTemplateData(r)

	data.Form = snippetCreateForm{
		Expires: 365,
	}

	a.render(w, r, http.StatusOK, "create.tmpl", data)
}

func (a *app) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var form snippetCreateForm

	err := a.decodePostForm(r, &form)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := a.newTemplateData(r)
		data.Form = form
		a.render(w, r, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	now := time.Now()
	var exp time.Time
	switch form.Expires {
	case 365:
		exp = now.AddDate(1, 0, 0)
	case 7:
		exp = now.AddDate(0, 0, 7)
	case 1:
		exp = now.AddDate(0, 0, 1)
	}
	ns := models.NewSnippet{
		Title:       form.Title,
		Content:     form.Content,
		DateExpires: exp,
	}

	// create a new snippet record in the database using the form data
	spt, err := a.snippets.Create(r.Context(), ns, now)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	// add flash message to the user session
	a.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	// redirect the user to the relevant page for the snippet.
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%s", spt.ID), http.StatusSeeOther)
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (a *app) userSignupForm(w http.ResponseWriter, r *http.Request) {
	data := a.newTemplateData(r)
	data.Form = userSignupForm{}
	a.render(w, r, http.StatusOK, "signup.tmpl", data)
}

func (a *app) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := a.decodePostForm(r, &form)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 10), "password", "This field must be at least 10 characters long")

	if !form.Valid() {
		data := a.newTemplateData(r)
		data.Form = form
		a.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	nu := models.NewUser{
		Email:    form.Email,
		Name:     form.Name,
		Password: form.Password,
	}
	_, err = a.users.Create(r.Context(), nu, time.Now())
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")

			data := a.newTemplateData(r)
			data.Form = form
			a.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			a.serverError(w, r, err)
		}

		return
	}

	a.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (a *app) loginUserForm(w http.ResponseWriter, r *http.Request) {
	data := a.newTemplateData(r)
	data.Form = userLoginForm{}
	a.render(w, r, http.StatusOK, "login.tmpl", data)
}

// userLoginPost checks the provided credentials and
// redirects the client to the requested path
func (a *app) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	err := a.decodePostForm(r, &form)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	// Check whether the credentials are valid. If they're not, add a generic error
	// message to the form failures map and re-display the login page.
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := a.newTemplateData(r)
		data.Form = form
		a.render(w, r, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	claims, err := a.users.Authenticate(r.Context(), time.Now(), form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrAuthenticationFailure) {
			form.AddNonFieldError("Email or password is incorrect")

			data := a.newTemplateData(r)
			data.Form = form
			a.render(w, r, http.StatusUnprocessableEntity, "login.tmpl", data)
		} else {
			a.serverError(w, r, err)
		}
		return
	}

	err = a.sessionManager.RenewToken(r.Context())
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	// Add the ID of the current user to the session data (user loged in)
	a.sessionManager.Put(r.Context(), "authenticatedUserID", claims.Subject)

	// pop the captured path from the session data
	path := a.sessionManager.PopString(r.Context(), "redirectPathAfterLogin")
	if path != "" {
		http.Redirect(w, r, path, http.StatusSeeOther)
		return
	}

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (a *app) logoutUserPost(w http.ResponseWriter, r *http.Request) {
	// remove authenticatedUserID from the session data (user logged out)
	a.sessionManager.Remove(r.Context(), "authenticatedUserID")
	// add flash message to the user session
	a.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *app) userProfile(w http.ResponseWriter, r *http.Request) {
	// get user ID from session data
	userID := a.sessionManager.GetString(r.Context(), "authenticatedUserID")

	// retreive user details from the database
	usr, err := a.users.QueryByID(r.Context(), userID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	//fmt.Fprintf(w, "%+v", user)

	data := a.newTemplateData(r)
	data.User = usr

	a.render(w, r, http.StatusOK, "profile.tmpl", data)
}

type changePasswordForm struct {
	CurrentPassword         string `form:"currentPassword"`
	NewPassword             string `form:"newPassword"`
	NewPasswordConfirmation string `form:"newPasswordConfirmation"`
	validator.Validator     `form:"-"`
}

func (a *app) changePasswordForm(w http.ResponseWriter, r *http.Request) {
	data := a.newTemplateData(r)
	data.Form = changePasswordForm{}
	a.render(w, r, http.StatusOK, "password.tmpl", data)
}

func (a *app) changePasswordPost(w http.ResponseWriter, r *http.Request) {
	var form changePasswordForm

	err := a.decodePostForm(r, &form)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.CurrentPassword), "currentPassword", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPassword), "newPassword", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPasswordConfirmation), "newPasswordConfirmation", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.NewPassword, 10), "newPassword", "This field must be at least 10 characters long")
	form.CheckField(validator.MinChars(form.NewPasswordConfirmation, 10), "newPasswordConfirmation", "This field must be at least 10 characters long")
	form.CheckField(validator.Equals(form.NewPassword, form.NewPasswordConfirmation), "newPassword", "This field must be equal to the new password confirmation")
	form.CheckField(!validator.Equals(form.NewPassword, form.CurrentPassword), "newPassword", "This field cannot be equal to the current password")

	if !form.Valid() {
		data := a.newTemplateData(r)
		data.Form = form
		a.render(w, r, http.StatusUnprocessableEntity, "password.tmpl", data)
		return
	}

	// get user ID from session data
	userID := a.sessionManager.GetString(r.Context(), "authenticatedUserID")

	// persist the password to the database
	err = a.users.ChangePassword(r.Context(), userID, form.CurrentPassword, form.NewPassword)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddFieldError("currentPassword", "Current password is incorrect")

			data := a.newTemplateData(r)
			data.Form = form
			a.render(w, r, http.StatusUnprocessableEntity, "password.tmpl", data)
		}

		a.serverError(w, r, err)
		return
	}

	// add flash message to the session data
	a.sessionManager.Put(r.Context(), "flash", "Your password has been updated!")
	// redirect browser to the users profile page
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}
