import * as request from 'superagent'

class Validator {
    public static CheckPassword(): void {
        var password = <HTMLInputElement>document.getElementById("password")
        var verify_password = <HTMLInputElement>document.getElementById("verify-password")
        if (password.value != verify_password.value) {
            verify_password.setCustomValidity("一致するパスワードを入力してください。");
        } else {
            verify_password.setCustomValidity('');
        }
    }
}

class Actions {
    private _token: string = ""
    constructor() {
        let e = <HTMLInputElement>document.getElementById('api-token')
        this._token = e.value
    }
    
    public get token(): string {
        return this._token
    }

    public DeleteUser(parent: JQuery<HTMLElement>): void {
        let id = parent.find("#user-id").val()
        request.post('/mypage/admin/api/delete')
        .set('Content-Type', 'application/json')
        .set('Authorization', `Bearer ${ this._token }`)
        .send({ id: id })
        .end(function(err, res){
            if (err || !res.ok) {
                alert('http error: ' + err);
            } else {
                let json = res.body
                if (json.is_success) {
                    alert('ユーザーを削除しました。')
                    window.location.reload(true)
                } else {
                    alert(`ユーザーの削除に失敗しました: ${ json.reason }`)
                    window.location.reload(true)
                }
            }
        })
    }
    
    public EditUser(parent: JQuery<HTMLElement>): void {
        let id = parent.find("#user-id").val()
        let is_admin: boolean = parent.find('#is-admin').is(':checked')
        request.post('/mypage/admin/api/edit')
            .set('Content-Type', 'application/json')
            .set('Authorization', `Bearer ${ this._token }`)
            .send({
                id: id,
                is_admin: is_admin
            })    
            .end(function(err, res){
                if (err || !res.ok) {
                    alert('http error: ' + err);
                } else {
                    let json = res.body
                    if (json.is_success) {
                        alert('ユーザーを編集しました。')
                        window.location.reload(true)
                    } else {
                        alert(`ユーザーの編集に失敗しました: ${ json.reason }`)
                        window.location.reload(true)
                    }
                }
            })

    }

    public CreateUser(parent: JQuery<HTMLElement>): void {
        let username = parent.find("#username").val()
        let password = parent.find("#password").val()
        let verify_password = parent.find("#verify-password").val()
        let is_admin: boolean = parent.find('#is-admin').is(':checked')
        console.log(is_admin)
        request.post('/mypage/admin/api/create')
            .set('Content-Type', 'application/json')
            .set('Authorization', `Bearer ${ this._token }`)
            .send({
                name: username,
                password: password,
                verify_password: verify_password,
                is_admin: is_admin
            })    
            .end(function(err, res){
                if (err || !res.ok) {
                    alert('http error: ' + err);
                } else {
                    let json = res.body
                    if (json.is_success) {
                        alert('ユーザーを作成しました。')
                        window.location.reload(true)
                    } else {
                        alert(`ユーザーの作成に失敗しました: ${ json.reason }`)
                        window.location.reload(true)
                    }
                }
            })
    }
}

$(document).on('submit', function(event) {
    $('form').find(':submit').prop('disabled', true);
})

$(document).ready(function() {
    $('#password').keyup(function() {
        Validator.CheckPassword()
    })
    $('#verify-password').keyup(function() {
        Validator.CheckPassword()
    })
})

$('#editModal').on('show.bs.modal', function (e) {
    let button = $(<HTMLElement>e.relatedTarget)
    let id = button.data('id')
    let name = button.data('name')
    let is_admin = button.data('is-admin')
    console.log(is_admin)
    let modal = $(this)
    modal.find('#username').val(name)
    modal.find('#user-id').val(id)
    modal.find('#is-admin').prop('checked', is_admin)
})

$('#deleteModal').on('show.bs.modal', function (e) {
    let button = $(<HTMLElement>e.relatedTarget)
    let id = button.data('id')
    let name = button.data('name')
    let modal = $(this)
    modal.find('#username').val(name)
    modal.find('#user-id').val(id)
})

var actions = new Actions()

var createElem = <HTMLInputElement>document.getElementById('create-user-validation')
createElem.addEventListener('submit', (e) => {
    e.preventDefault()
    actions.CreateUser($("#create-user-validation"))
    console.log(actions.token)
})

var editElem = <HTMLInputElement>document.getElementById('edit-user-validation')
editElem.addEventListener('submit', (e) => {
    e.preventDefault()
    actions.EditUser($("#edit-user-validation"))
    console.log(actions.token)
})

var deleteElem = <HTMLInputElement>document.getElementById('delete-user-validation')
deleteElem.addEventListener('submit', (e) => {
    e.preventDefault()
    actions.DeleteUser($("#delete-user-validation"))
    console.log(actions.token)
})