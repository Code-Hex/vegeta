import * as request from 'superagent';

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
    
    public CreateUser(): void {
        let username = $("#username").val()
        let password = $("#password").val()
        let verify_password = $("#verify-password").val()
        let is_admin: boolean = $('#is-admin').is(':checked')
        console.log(is_admin)
        request.post('/api/create')
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
                        alert(`ユーザーの作成に失敗しました: ${ json.reason }`);
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

var actions = new Actions()
var createElem = <HTMLInputElement>document.getElementById('create-user-validation')
createElem.addEventListener('submit', (e) => {
    e.preventDefault()
    console.log("executed")
    actions.CreateUser()    
    console.log(actions.token)
})