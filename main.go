package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type handler struct {
	ugh *UghRM
}

func die(w http.ResponseWriter, status int, message string) bool {
	w.WriteHeader(status)
	w.Write([]byte(fmt.Sprintf(`{"error": %q}`, message)))
	return true
}

func marshal(w http.ResponseWriter, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("error marshaling: %s", err)
		die(w, http.StatusInternalServerError, "error marshaling")
		return
	}
	w.Write(b)
}

func (h *handler) ListTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		die(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	rs := h.ugh.All("todos", 1)
	if rs.IsErr() {
		log.Printf("error listing todos: %s", rs.Err())
		die(w, http.StatusInternalServerError, "error listening todos")
		return
	}

	marshal(w, rs.OK())
}

func (h *handler) NewTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		die(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		die(w, http.StatusBadRequest, "need a title")
		return
	}

	todo := map[string]interface{}{
		"title": title,
		"done":  false,
	}

	rs := h.ugh.Insert("todos", todo)
	if rs.IsErr() {
		log.Printf("error inserting todos: %s", rs.Err())
		die(w, http.StatusInternalServerError, "error inserting todo")
		return
	}

	todo["id"] = rs.OK()
	marshal(w, todo)
}

func (h *handler) DoneTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		die(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		die(w, http.StatusNotFound, "not found")
		return
	}
	rs := h.ugh.Get("todos", id)
	if rs.IsErr() {
		die(w, http.StatusNotFound, "not found")
		return
	}
	todo := rs.OK()

	d := false
	done := r.FormValue("done")
	if done == "true" || done == "1" {
		d = true
	}

	if rs := h.ugh.Update("todos", id, map[string]interface{}{
		"done": d,
	}); !rs.IsErr() {
		todo["done"] = d
		marshal(w, todo)
		return
	}

	die(w, http.StatusInternalServerError, "db error")
}

func (h *handler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		die(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil || h.ugh.Delete("todos", id).IsErr() {
		die(w, http.StatusNotFound, "not found")
		return
	}
	w.Write([]byte("{}"))
}

func openDB(dbURL string) *Result[*sql.DB] {
	return Lift(sql.Open("postgres", dbURL))
}

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("SQL: error opening db: %q", err)
	}

	/* Just for some more customization */
	user := os.Getenv("USER")
	if user == "" { user = "user" }
	quote := os.Getenv("QUOTE")
	connectedAs := os.Getenv("CONNECTED_AS")
	index = strings.ReplaceAll(index, "%%USER%%", user)
	index = strings.ReplaceAll(index, "%%QUOTE%%", quote)
	index = strings.ReplaceAll(index, "%%CONNECTED_AS%%", connectedAs)

	ugh := NewUghRM(db)
	api := &handler{ugh: ugh}

	http.HandleFunc("/api/list", api.ListTodos)
	http.HandleFunc("/api/new", api.NewTodo)
	http.HandleFunc("/api/done", api.DoneTodo)
	http.HandleFunc("/api/delete", api.DeleteTodo)
	// serve the static file.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(index))
	})
	log.Printf("here we go!")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

var index = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>TODOs on Tripods</title>
    <style>
      body {
       text-align: center;
       font-family: Helvetica, sans;
      }
      header {
       color: #333;
      }
      header h1 span {
       font-style: italic;
      }
      main, header, footer {
       width: 600px;
       margin: 0 auto;
       text-align: left;
      }
      footer em {
       color: #111;
      }
      #todos, #todo-new {
       margin: 0;
       padding: 0;
      }
      #todos li, #todo-new li {
       font-size: 24px;
       list-style-type: none;
       line-height: 1.5;
      }
      #todos li input[type="checkbox"], #todo-new li input[type="checkbox"] {
       vertical-align: middle;
       width: 20px;
       height: 20px;
      }
      #todos li span {
       margin: 0 10px;
      }
      #todos li button {
       font-size: 10px;
       color: red;
       background: none;
       border: 0;
       vertical-align: middle;
      }
      #todo-new {
       border-top: 1px solid #ccc;
       margin-top: 10px;
       padding-top: 10px;
      }
      #todo-new li input[type="text"] {
       font-size: 24px;
      }
      .strikeout {
       text-decoration: line-through;
       font-style: italic;
      }
      footer {
       margin-top: 100px;
       font-size: 16px;
       color: #aaa;
      }
    </style>
  </head>
  <body>
    <header>
      <h1>%%USER%%'s TODOs (not really) on <span>Tripods</span></h1>
    </header>
    <main>
      <ul id="todos">

      </ul>
      <ul id="todo-new">
        <li>
          <input type="checkbox" disabled="true" class="todo-done" />
          <input type="text" id="todo-text-new" />
          <button id="todo-button-new">Add</button>
        </li>
      </ul>
    </main>
    <footer>
      &copy; 2005 73 Sinewaves, LLC. &mdash; <q>%%QUOTE%%</q> &mdash; <em>connected as %%CONNECTED_AS%%</em>

    </footer>
  </body>
  <script>
    var makeRequest = function(method, url, body, callback) {
        var r = new XMLHttpRequest();
        r.open(method, url, true);
        r.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
        r.onreadystatechange = function() {
            if (this.readyState == 4) {
                if (this.status != 200) {
                    callback(null, {"status": this.status});
                } else {
                    callback(JSON.parse(r.response), null);
                }
            }
        };
        r.send(body);
    };
    var doPost = function(url, params, callback) {
        var data = [];
        for (p in params) {
            data.push(encodeURIComponent(p) + "=" + encodeURIComponent(params[p]));
        }
        makeRequest("POST", url, data.join("&"), callback);
    };

    var doGet = function(url, callback) {
        console.log(url, "here we go");
        makeRequest("GET", url, null, callback);
    };

    var goList = function(callback) {
        console.log("HELLO");
        // var todos = [
        //     { "id": 1
        //     , "title": "A todo"
        //     , "done": false
        //     },
        //     { "id": 2
        //       , "title": "Some other todo"
        //       , "done": false
        //     },
        //     { "id": 3
        //       , "title": "A done todo"
        //       , "done": true
        //     }
        // ];
        doGet("/api/list", callback);
        // callback(todos, null);
    };

    var goDelete = function(id, callback) {
        // actually do the delete, of course.
        doPost("/api/delete", {"id": id}, callback);
        // callback({}, null);
    };

    var goNew = function(title, callback) {
        // actually make the call.
        doPost("/api/new", {"title": title, "done": false}, callback);
        // callback({"id": Math.floor(Math.random() * 100)
        //         , "title": title
        //         , "done": false
        //         }, null);
    };

    var goMark = function(id, done, callback) {
        doPost("/api/done", {"id": id, "done": done}, callback);
        // callback({ "id": id
        //         , "done": done
        //         , "title": ""
        //         }, null);
    };


    var installTODO = function(parent, todo) {
        const li = document.createElement("li");
        li.dataset.id = todo.id;
        li.className = "";
        const label = document.createElement("label");
        label.setAttribute("for", "box-done-" + todo.id);
        const checkbox = document.createElement("input");
        checkbox.setAttribute("type", "checkbox");
        checkbox.dataset.id = todo.id;
        checkbox.id = "box-done-" + todo.id;
        checkbox.className = "todo-done";
        checkbox.checked = todo.done;
        checkbox.addEventListener("change", function(e) {
            // need to make a request to mark this as done or not done.
            var target = e.currentTarget;
            goMark(target.dataset.id, target.checked, function(result, error) {
                if (error === null) {
                    if (result.done) {
                        text.className = "strikeout";
                    } else {
                        text.className = "";
                    }
                }
            });
            e.stopPropagation();
        });
        const text = document.createElement("span");
        text.innerText = todo.title;
        if (todo.done) {
            text.className = "strikeout";
        }
        const button = document.createElement("button");
        button.innerText = "X";
        button.dataset.id = todo.id;
        button.addEventListener("click", function(e) {
            var target = e.currentTarget;
            goDelete(target.dataset.id, function(result, error) {
                if (error === null) {
                    target.parentNode.remove();
                }
            });
        });

        label.appendChild(checkbox);
        label.appendChild(text);
        label.appendChild(button);
        li.appendChild(label);
        parent.appendChild(li);
    };

    var setup = function() {
        goList(function(todos, error) {
            console.log(todos);
            console.log(error);
            var parent = document.getElementById("todos");
            for (var i = 0; i < todos.length; i++) {
                installTODO(parent, todos[i]);
            }
            console.log("done");

        });

        var addTODOEvent = function(e) {
            e.preventDefault();
            var title = document.getElementById("todo-text-new");
            if (title.value != "") {
                goNew(title.value, function(todo, error) {
                    installTODO(document.getElementById("todos"), todo);
                    title.value = "";
                });
            }
        }

        // Add a TODO.
        document.getElementById("todo-button-new").addEventListener("click", addTODOEvent);
        document.getElementById("todo-text-new").addEventListener("keyup", function(e) {
            if (e.keyCode == 13) {
                addTODOEvent(e);
            }
        });
    };
    document.addEventListener('DOMContentLoaded', (e) => {
        setup();
    });
  </script>
</html>
`
