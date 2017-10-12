import * as request from 'superagent';

class Settings {
    private _token: string = ""
    constructor() {
        let e = <HTMLInputElement>document.getElementById('api-token')
        this._token = e.value
    }

    public RegenerateToken(): void {
        request.patch('/mypage/api/regenerate')
        .set('Content-Type', 'application/json')
        .set('Authorization', `Bearer ${ this._token }`)
        .send()
        .end(function(err, res){
            if (err || !res.ok) {
                alert('http error: ' + err);
            } else {
                let json = res.body
                if (json.is_success) {
                    alert('アクセストークンを更新しました')
                    window.location.reload(true)
                } else {
                    alert(`${ json.reason }`)
                    window.location.reload(true)
                }
            }
        })
    }
    
    public RegisterPassword(): void {
        let passwdElem = <HTMLInputElement>document.getElementById('password')
        let passwdVerifyElem = <HTMLInputElement>document.getElementById('password-verify')
        let password = passwdElem.value
        let password_verify = passwdVerifyElem.value
        if (password == "" || password_verify == "") {
            alert("パスワードが入力されていません")
            return;
        }
        if (password != password_verify) {
            alert("パスワードが一致していません")
            return;
        }
        request.post('/mypage/api/reregister_password')
        .set('Content-Type', 'application/json')
        .set('Authorization', `Bearer ${ this._token }`)
        .send({ password: password, verify_password: password_verify })
        .end(function(err, res){
            if (err || !res.ok) {
                alert('http error: ' + err);
            } else {
                let json = res.body
                if (json.is_success) {
                    alert('パスワードを更新しました')
                    window.location.reload(true)
                } else {
                    alert(`パスワードの更新に失敗しました: ${ json.reason }`)
                    window.location.reload(true)
                }
            }
        })
    }
}

var settings = new Settings()

var regenElem = <HTMLInputElement>document.getElementById('regen-token')
regenElem.addEventListener('click', (e) => {
    e.preventDefault()
    settings.RegenerateToken()
})

var reregister = <HTMLInputElement>document.getElementById("reregister-password")
reregister.addEventListener('click', (e) => {
    e.preventDefault()
    settings.RegisterPassword()
})