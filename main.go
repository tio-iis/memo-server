package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("start")
	http.HandleFunc("/", hanlder)
	http.ListenAndServe(":8080", nil)
}

func hanlder(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, htmlStr)
}

const htmlStr = `
<!doctype html>
<html lang="en">

<head>
  <!-- Required meta tags -->
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">

  <!-- Bootstrap CSS -->
  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.0/dist/css/bootstrap.min.css" rel="stylesheet"
    integrity="sha384-KyZXEAg3QhqLMpG8r+8fhAXLRk2vvoC2f3B09zVXn8CA5QIVfZOJ3BCsw2P0p/We" crossorigin="anonymous">

  <script src="https://unpkg.com/dayjs@1.8.21/dayjs.min.js"></script>

  <title>メモ帳アプリ</title>
</head>

<body style="padding: 32px; background-color: rgb(248, 249, 250); color: rgb(33, 37, 41);">
  <div>
    <h2>メモ帳アプリ</h3>
      <p>自分のメモを管理できるアプリです。</p>
  </div>
  <div style="padding-bottom: 24px;">
    <button id="id-add-button" class="btn btn-primary">メモを登録する</button>
  </div>
  <div>
    <div style="padding-bottom: 24px;">
      <button id="id-delete-button" class="btn btn-secondary">削除する</button>
    </div>
    <table class="table table-hover">
      <thead>
        <tr>
          <th scope="col"><input type="checkbox" id="id-delete-all-memos"></th>
          <th scope="col">タイトル</th>
          <th scope="col">作成日</th>
          <th scope="col">更新日</th>
          <th scope="col">編集</th>
        </tr>
      </thead>
      <tbody id="id-memo-list">
      </tbody>
    </table>


    <div class="modal fade" id="editModal" tabindex="-1" aria-labelledby="editModalLabel" aria-hidden="true">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="editModalLabel">編集</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <div id="id-validation-error-in-edit-modal" class="alert alert-danger" style="display:none;" role="alert">
            </div>
            <div class="input-group mb-3">
              <span class="input-group-text" id="inputGroup-sizing-default">タイトル</span>
              <input type="text" id="id-modal-title" class="form-control" aria-label="Sizing example input"
                aria-describedby="inputGroup-sizing-default">
              <input type="hidden" id="id-memo-id">
            </div>
            <div class="form-floating">
              <textarea class="form-control" style="height: 150px;" placeholder="Leave a comment here"
                id="id-modal-body"></textarea>
              <label for="id-modal-body">本文</label>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
            <button type="button" id="id-update-button" class="btn btn-primary">更新</button>
          </div>
        </div>
      </div>
    </div>

    <div class="modal fade" id="id-add-modal" tabindex="-1" aria-labelledby="editModalLabel" aria-hidden="true">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="id-add-modal-label">登録</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <div id="id-validation-error-in-add-modal" class="alert alert-danger" style="display:none;" role="alert">
            </div>
            <div class="input-group mb-3">
              <span class="input-group-text" id="inputGroup-sizing-default">タイトル</span>
              <input type="text" id="id-add-modal-title" class="form-control" aria-label="Sizing example input"
                aria-describedby="inputGroup-sizing-default">
            </div>
            <div class="form-floating">
              <textarea class="form-control" style="height: 150px;" placeholder="Leave a comment here"
                id="id-add-modal-body"></textarea>
              <label for="id-add-modal-body">本文</label>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
            <button type="button" id="id-add-modal-button" class="btn btn-primary">登録</button>
          </div>
        </div>
      </div>
    </div>

  </div>


  <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.0/dist/js/bootstrap.bundle.min.js"
    integrity="sha384-U1DAWAznBHeqEIlVSCgzq+c9gqGAJn5c/t99JyeKa9xxaYpSvHU5awsuZVVFIhvj"
    crossorigin="anonymous"></script>
</body>

<script>


  const editModal = new bootstrap.Modal(document.getElementById('editModal'))
  const addModal = new bootstrap.Modal(document.getElementById('id-add-modal'))
  const titleInAddModal = document.getElementById("id-add-modal-title")
  const bodyInAddModal = document.getElementById("id-add-modal-body")
  const memoList = document.getElementById("id-memo-list")
  const validationErrorInAddModal = document.getElementById("id-validation-error-in-add-modal")
  const validationErrorInEditModal = document.getElementById("id-validation-error-in-edit-modal")
  const memoIdInEditModal = document.getElementById("id-memo-id")
  const titleInEditModal = document.getElementById("id-modal-title")
  const bodyInEditModal = document.getElementById("id-modal-body")
  const checkboxes = document.getElementsByClassName(getCheckboxClass())

  function getIdTitleInMemoList(memoId) {
    return document.getElementById(createTitleInMemoListId(memoId))
  }

  function createTitleInMemoListId(memoId) {
    return "id-title-in-list-" + memoId
  }

  function createUpdatedAtId(memoId) {
    return "id-updated-at-" + memoId
  }

  function getIdBodyInlist(memoId) {
    return document.getElementById(createBodyInListId(memoId))
  }

  function createBodyInListId(memoId) {
    return "id-body-in-list-" + memoId
  }

  function getCheckboxClass() {
    return "class-checkbox"
  }

  function validateMemo(title, body, errorHTMLElement, memoId) {
    errorHTMLElement.innerHTML = ""
    if (title.length < 1 || title.length > 30) {
      errorHTMLElement.innerHTML = "タイトルの文字数は1文字以上30文字以下にしてください。<br>"
    }

    if (body.length < 1 || body.length > 100) {
      errorHTMLElement.innerHTML = errorHTMLElement.innerHTML + "本文の文字数は1文字以上100文字以下にしてください。<br>"
    }

    if (memoList.children.length > 0) {
      Array.from(memoList.children).forEach((tr) => {
        if (tr.id == memoId) {
          return
        }

        const t = getIdTitleInMemoList(tr.id).innerText
        if (title == t) {
          errorHTMLElement.innerHTML = errorHTMLElement.innerHTML + "すでに登録済みのタイトルです。<br>"
        }
      })
    }

    if (errorHTMLElement.innerHTML.length > 0) {
      errorHTMLElement.style.display = ""
      return false
    }

    return true
  }

  const addButtonInModal = document.getElementById("id-add-modal-button")
  addButtonInModal.addEventListener("click", (event) => {
    const title = titleInAddModal.value
    const body = bodyInAddModal.value

    if (validateMemo(title, body, validationErrorInAddModal, "0") == false) {
      return
    }

    const now = dayjs()
    const createdAt = now.format("YYYY-MMM-DD HH:mm:ss")
    const updatedAt = "なし"
    const memoId = now.valueOf()

    const tr = document.createElement("tr")
    tr.setAttribute("id", memoId)

    const editButtonId = "id-edit-button-" + memoId
    const idTitleInList = createTitleInMemoListId(memoId)
    const updatedAtId = createUpdatedAtId(memoId)

    tr.innerHTML = '<td><input class="' + getCheckboxClass() + '" data-memo-id="' + memoId + '" type="checkbox"></td><td id="' + idTitleInList + '">' + title + '</td><td>' + createdAt + '</td><td id="' + updatedAtId + '">' + updatedAt + '</td><td><button type="button" data-memo-id="' + memoId + '" id="' + editButtonId + '" class="btn btn-primary">編集</button></td>'
    memoList.appendChild(tr)

    const inputForBody = document.createElement("input")
    inputForBody.setAttribute("id", createBodyInListId(memoId))
    inputForBody.setAttribute("type", "hidden")
    inputForBody.value = body
    tr.appendChild(inputForBody)

    const editButton = document.getElementById(editButtonId)
    editButton.addEventListener("click", (event) => {
      validationErrorInEditModal.style.display = "none"
      validationErrorInEditModal.innerHTML = ""

      const memoId = event.currentTarget.dataset.memoId
      memoIdInEditModal.value = memoId

      const title = getIdTitleInMemoList(memoId).innerText
      titleInEditModal.value = title

      const body = getIdBodyInlist(memoId).value
      bodyInEditModal.value = body

      editModal.show()
    })
    addModal.hide()
  })

  const addButton = document.getElementById("id-add-button")
  addButton.addEventListener("click", (event) => {
    titleInAddModal.value = ""
    bodyInAddModal.value = ""
    addModal.show()

    validationErrorInAddModal.style.display = "none"
    validationErrorInAddModal.innerHTML = ""
  })

  const deleteButton = document.getElementById("id-delete-button")
  deleteButton.addEventListener("click", (event) => {
    Array.from(checkboxes).forEach((checkbox) => {
      if (checkbox.checked == false) {
        return false
      }

      const memoId = checkbox.dataset.memoId
      document.getElementById(memoId).remove()
    })
  })

  const deleteAllMemos = document.getElementById("id-delete-all-memos")
  deleteAllMemos.addEventListener("change", (event) => {
    if (checkboxes.length == 0) {
      return
    }

    Array.from(checkboxes).forEach((checkbox) => {
      if (deleteAllMemos.checked == true) {
        checkbox.checked = true
      } else {
        checkbox.checked = false
      }
    })
  })

  const updateButton = document.getElementById("id-update-button")
  updateButton.addEventListener("click", (event) => {
    const memoId = memoIdInEditModal.value
    const title = titleInEditModal.value
    const body = bodyInEditModal.value

    if (validateMemo(title, body, validationErrorInEditModal, memoId) == false) {
      return
    }

    //メモリストのタイトル
    getIdTitleInMemoList(memoId).innerText = title

    //本文
    getIdBodyInlist(memoId).value = body

    document.getElementById(createUpdatedAtId(memoId)).innerText = dayjs().format("YYYY-MM-DD HH:mm:ss")

    editModal.hide()
  })

</script>

</html>
`