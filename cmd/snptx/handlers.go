package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tullo/snptx/internal/forms"
	"github.com/tullo/snptx/internal/snippet"
	"github.com/tullo/snptx/internal/user"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	s, err := app.snippets.Latest(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "home.page.tmpl", &templateData{
		Snippets: s,
	})
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "about.page.tmpl", &templateData{})
}

func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	// pat does not strip the colon from the named capture key,
	// get the value of ":id" from the query string instead of "id"
	id := r.URL.Query().Get(":id")
	s, err := app.snippets.Retrieve(r.Context(), id)
	if err != nil {
		// unwrapping errors
		if errors.Is(err, snippet.ErrNotFound) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.render(w, r, "show.page.tmpl", &templateData{
		Snippet: s,
	})
}

func (app *application) updateSnippetForm(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")
	s, err := app.snippets.Retrieve(r.Context(), id)
	if err != nil {
		// unwrapping errors
		if errors.Is(err, snippet.ErrNotFound) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := make(map[string][]string)
	data["title"] = append(data["title"], s.Title)
	data["content"] = append(data["content"], s.Content)

	app.render(w, r, "edit.page.tmpl", &templateData{
		Snippet: &snippet.Info{ID: id},
		Form:    forms.New(data),
	})
}

func (app *application) updateSnippet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := forms.New(r.PostForm)
	form.Required("title", "content")
	form.MaxLength("title", 100)
	if !form.Valid() {
		app.render(w, r, "edit.page.tmpl", &templateData{
			Snippet: &snippet.Info{ID: id},
			Form:    form})
		return
	}
	// update snippet record in the database using the form data
	t := form.Get("title")
	c := form.Get("content")
	up := snippet.UpdateSnippet{
		Title:   &t,
		Content: &c,
	}

	err = app.snippets.Update(r.Context(), id, up, time.Now())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r, "flash", "Snippet successfully updated!")
	http.Redirect(w, r, fmt.Sprintf("/snippet/%s", id), http.StatusSeeOther)
}

func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {
	// render the form using an empty forms.Form struct
	app.render(w, r, "create.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {

	// add data in POST request body to the r.PostForm map
	err := r.ParseForm()
	if err != nil {
		// no body, or body is too large to process
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// create a new forms.Form struct containing the POSTed data,
	// then use the validation methods to check the content.
	form := forms.New(r.PostForm)
	form.Required("title", "content", "expires")
	form.MaxLength("title", 100)
	form.PermittedValues("expires", "365", "7", "1")

	// if the form is not valid, redisplay the template passing in the parsed data.
	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{Form: form})
		return
	}

	now := time.Now()
	var exp time.Time
	switch form.Get("expires") {
	case "365":
		exp = now.AddDate(1, 0, 0)
	case "7":
		exp = now.AddDate(0, 0, 7)
	case "1":
		exp = now.AddDate(0, 0, 7)
	}
	ns := snippet.NewSnippet{
		Title:       form.Get("title"),
		Content:     form.Get("content"),
		DateExpires: exp,
	}

	// create a new snippet record in the database using the form data
	spt, err := app.snippets.Create(r.Context(), ns, now)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// add flash message to the user session
	app.session.Put(r, "flash", "Snippet successfully created!")

	// redirect the user to the relevant page for the snippet.
	http.Redirect(w, r, fmt.Sprintf("/snippet/%s", spt.ID), http.StatusSeeOther)
}

func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MaxLength("name", 255)
	form.MaxLength("email", 255)
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 10)

	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	}

	nu := user.NewUser{
		Email:    form.Get("email"),
		Name:     form.Get("name"),
		Password: form.Get("password"),
	}
	_, err = app.users.Create(r.Context(), nu, time.Now())
	if err != nil {
		if errors.Is(err, user.ErrDuplicateEmail) {
			form.Errors.Add("email", "Address is already in use")
			app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.session.Put(r, "flash", "Your signup was successful. Please log in.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

// loginUser checks the provided credentials and redirects the client to the
// requested path
func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Check whether the credentials are valid. If they're not, add a generic error
	// message to the form failures map and re-display the login page.
	form := forms.New(r.PostForm)
	claims, err := app.users.Authenticate(r.Context(), time.Now(), form.Get("email"), form.Get("password"))
	if err != nil {
		if errors.Is(err, user.ErrAuthenticationFailure) {
			form.Errors.Add("generic", "Email or Password is incorrect")
			app.render(w, r, "login.page.tmpl", &templateData{Form: form})
			return
		}
		app.serverError(w, err)
		return
	}

	// Add the ID of the current user to the session data (user loged in)
	app.session.Put(r, "authenticatedUserID", claims.Subject)

	// pop the captured path from the session data
	path := app.session.PopString(r, "redirectPathAfterLogin")
	if path != "" {
		http.Redirect(w, r, path, http.StatusSeeOther)
		return
	}

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	// remove authenticatedUserID from the session data (user logged out)
	app.session.Remove(r, "authenticatedUserID")
	// add flash message to the user session
	app.session.Put(r, "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userProfile(w http.ResponseWriter, r *http.Request) {
	// get user ID from session data
	userID := app.session.GetString(r, "authenticatedUserID")

	// retreive user details from the database
	usr, err := app.users.Retrieve(r.Context(), userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	//fmt.Fprintf(w, "%+v", user)
	app.render(w, r, "profile.page.tmpl", &templateData{
		User: usr,
	})

}

func (app *application) changePasswordForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "password.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) changePassword(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("currentPassword", "newPassword", "newPasswordConfirmation")
	form.MinLength("newPassword", 10)
	form.MinLength("newPasswordConfirmation", 10)
	if form.Get("newPassword") != form.Get("newPasswordConfirmation") {
		form.Errors.Add("newPasswordConfirmation", "Passwords do not match")
	}
	if form.Get("currentPassword") == form.Get("newPassword") {
		form.Errors.Add("newPassword", "Your new password must not match your previous")
	}

	if !form.Valid() {
		app.render(w, r, "password.page.tmpl", &templateData{Form: form})
		return
	}

	// get user ID from session data
	userID := app.session.GetString(r, "authenticatedUserID")

	// persist the password to the database
	err = app.users.ChangePassword(r.Context(), userID, form.Get("currentPassword"), form.Get("newPassword"))
	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			form.Errors.Add("currentPassword", "Current password is incorrect")
			app.render(w, r, "password.page.tmpl", &templateData{Form: form})
		} else if err != nil {
			app.serverError(w, err)
		}
		return
	}

	// add flash message to the session data
	app.session.Put(r, "flash", "Your password has been updated!")
	// redirect browser to the users profile page
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}
