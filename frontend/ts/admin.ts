class Validator {
    public static CheckPassword() : void {
        var password = <HTMLInputElement>document.getElementById("password")
        var verify_password = <HTMLInputElement>document.getElementById("verify-password")

        if (password.value != verify_password.value) {
            verify_password.setCustomValidity("Passwords Don't Match");
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