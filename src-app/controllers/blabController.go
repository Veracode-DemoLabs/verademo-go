package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"verademo-go/src-app/models"
	sqlite "verademo-go/src-app/shared/db"
	session "verademo-go/src-app/shared/session"
	"verademo-go/src-app/shared/view"

	"log"
)

var sqlBlabsByMe = `SELECT blabs.content, blabs.timestamp, COUNT(comments.blabber), blabs.blabid ` +
	`FROM blabs LEFT JOIN comments ON blabs.blabid = comments.blabid ` +
	`WHERE blabs.blabber = ? GROUP BY blabs.blabid ORDER BY blabs.timestamp DESC;`

var sqlBlabsForMe = `SELECT users.username, users.blab_name, blabs.content, blabs.timestamp, COUNT(comments.blabber), blabs.blabid ` +
	`FROM blabs INNER JOIN users ON blabs.blabber = users.username INNER JOIN listeners ON blabs.blabber = listeners.blabber ` +
	`LEFT JOIN comments ON blabs.blabid = comments.blabid WHERE listeners.listener = ? ` +
	`GROUP BY blabs.blabid ORDER BY blabs.timestamp DESC LIMIT %d OFFSET %d;`

func ShowFeed(w http.ResponseWriter, r *http.Request) {

	// Struct for variables to pass to the feed template
	type Outputs struct {
		BlabsByOthers []models.Blab
		BlabsByMe     []models.Blab
		CurrentUser   string
		Error         string
	}

	// Check session username
	sess := session.Instance(r)

	if sess.Values["username"] == nil {
		log.Println("User is not Logged In - redirecting...")
		http.Redirect(w, r, "login?target=feed", http.StatusFound)
		return
	}

	username := sess.Values["username"].(string)

	log.Println("User is Logged In - continuing... UA=" + r.Header.Get("user-agent") + " U=" + username)

	var outputs Outputs

	// Get blabs from blabbers that are being listened to
	log.Println("Executing query to get all 'Blabs for me'")
	blabsForMe := fmt.Sprintf(sqlBlabsForMe, 10, 0)
	blabsForMeResults, err := sqlite.DB.Query(blabsForMe, username)
	if err != nil {
		errMsg := "Error getting 'Blabs for me':\n" + err.Error()
		log.Println(errMsg)
		outputs.Error = errMsg
		view.Render(w, "feed.html", outputs)
		return
	}

	// Close the results object when they have been used up
	defer blabsForMeResults.Close()

	// Add each blab found to a variable to be passed to the template
	var feedBlabs []models.Blab

	for blabsForMeResults.Next() {
		var author models.Blabber
		var post models.Blab

		if err := blabsForMeResults.Scan(&author.Username, &author.BlabName, &post.Content, &post.PostDate, &post.CommentCount, &post.Id); err != nil {
			errMsg := "Error reading data from 'Blabs for me' query:\n" + err.Error()
			log.Println(errMsg)
			outputs.Error = errMsg
			view.Render(w, "feed.html", outputs)
			return
		}

		post.Author = author
		post.PostDate = models.Timestamp(post.PostDate)

		feedBlabs = append(feedBlabs, post)

	}

	outputs.BlabsByOthers = feedBlabs
	outputs.CurrentUser = username

	// Get blabs from the current user
	log.Println("Executing query to get all of user's Blabs")
	blabsByMeResults, err := sqlite.DB.Query(sqlBlabsByMe, username)
	if err != nil {
		errMsg := "Error getting 'Blabs for me':\n" + err.Error()
		log.Println(errMsg)
		outputs.Error = errMsg
		view.Render(w, "feed.html", outputs)
		return
	}

	// Close the results object when they have been used up
	defer blabsByMeResults.Close()

	// Add each blab found to a variable to be passed to the template
	var myBlabs []models.Blab

	for blabsByMeResults.Next() {
		var post models.Blab

		if err := blabsByMeResults.Scan(&post.Content, &post.PostDate, &post.CommentCount, &post.Id); err != nil {
			errMsg := "Error reading data from 'Blabs by me' query:\n" + err.Error()
			log.Println(errMsg)
			outputs.Error = errMsg
			view.Render(w, "feed.html", outputs)
			return
		}

		post.PostDate = models.Timestamp(post.PostDate)

		myBlabs = append(myBlabs, post)

	}

	outputs.BlabsByMe = myBlabs

	view.Render(w, "feed.html", outputs)

}

func MoreFeed(w http.ResponseWriter, r *http.Request) {
	countParam := r.URL.Query().Get("count")
	lenParam := r.URL.Query().Get("len")

	// Template for response
	template := "<li><div>" + "\t<div class=\"commenterImage\">" + "\t\t<img src=\"/static/images/%s.png\">" +
		"\t</div>" + "\t<div class=\"commentText\">" + "\t\t<p>%s</p>" +
		"\t\t<span class=\"date sub-text\">by %s on %s</span><br>" +
		"\t\t<span class=\"date sub-text\"><a href=\"blab?blabid=%d\">%d Comments</a></span>" + "\t</div>" +
		"</div></li>"

	// Convert GET parameters to integers
	count, err := strconv.Atoi(countParam)
	if err != nil {
		log.Println("Error converting count:" + countParam + " to integer:\n" + err.Error())
		http.Redirect(w, r, "feed", http.StatusBadRequest)
		return
	}

	len, err := strconv.Atoi(lenParam)
	if err != nil {
		log.Println("Error converting len:" + lenParam + " to integer:\n" + err.Error())
		http.Redirect(w, r, "feed", http.StatusBadRequest)
		return
	}

	// Check session username
	sess := session.Instance(r)

	if sess.Values["username"] == nil {
		log.Println("User is not Logged In - redirecting...")
		http.Redirect(w, r, "login?target=feed", http.StatusFound)
		return
	}

	username := sess.Values["username"].(string)

	// Run SQL query
	log.Println("Executing query to get more blabs")
	blabsForMe := fmt.Sprintf(sqlBlabsForMe, len, count)
	results, err := sqlite.DB.Query(blabsForMe, username)
	if err != nil {
		errMsg := "Error getting more blabs:\n" + err.Error()
		log.Println(errMsg)
		http.Redirect(w, r, "feed", http.StatusBadRequest)
		return
	}

	// Close the results object when they have been used up
	defer results.Close()

	// Add each blab found to the response using the template
	var ret string

	for results.Next() {
		var author models.Blabber
		var post models.Blab

		if err := results.Scan(&author.Username, &author.BlabName, &post.Content, &post.PostDate, &post.CommentCount, &post.Id); err != nil {
			errMsg := "Error reading data from 'more feed' query:\n" + err.Error()
			log.Println(errMsg)
			http.Redirect(w, r, "feed", http.StatusBadRequest)
			return
		}

		ret += fmt.Sprintf(template, author.Username, post.Content, author.BlabName, models.Timestamp(post.PostDate), post.Id, post.CommentCount)

	}

	// Write the response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	_, err = w.Write([]byte(ret))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}

}
