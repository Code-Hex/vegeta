class Validator {
    public static CheckPassword() : void {
        var password = <HTMLInputElement>document.getElementById("password")
        var verify_password = <HTMLInputElement>document.getElementById("verify-password")
        if (password.value != verify_password.value) {
            verify_password.setCustomValidity("一致するパスワードを入力してください。");
        } else {
            verify_password.setCustomValidity('');
        }
    }
}

$(document).ready(function() {
    $('#password').keyup(function() {
        Validator.CheckPassword()
    })
    $('#verify-password').keyup(function() {
        Validator.CheckPassword()
    })
})