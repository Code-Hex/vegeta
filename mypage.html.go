// Code generated by hero.
// source: /Users/codehex/Desktop/go/src/github.com/Code-Hex/vegeta/template/mypage.html
// DO NOT EDIT!
package vegeta

import (
	"io"

	"github.com/shiyanhui/hero"
)

func MyPageHTML(args Args, w io.Writer) {
	_buffer := hero.GetBuffer()
	defer hero.PutBuffer(_buffer)
	_buffer.WriteString(`<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="IoTを用いた栽培中の植物のデータを管理するプロジェクトです。">
  <link href="/assets/css/main.css" rel="stylesheet">
  <link href="https://maxcdn.bootstrapcdn.com/font-awesome/4.7.0/css/font-awesome.min.css" rel="stylesheet" integrity="sha384-wvfXpqpZZVQGK6TAh5PVlGOfQNHSoD2xbE+QkPxCAFlNEevoEH3Sl0sibVcOQVnN" crossorigin="anonymous">
  <link rel="stylesheet" href="/assets/css/bootstrap.css">
  <script src="/assets/js/jquery.min.js"></script>
  <script src="/assets/js/tether.min.js"></script>
  <script src="/assets/js/bootstrap.min.js"></script>
  `)
	_buffer.WriteString(`
  <link rel="stylesheet" type="text/css" href="https://cdn.datatables.net/1.10.16/css/dataTables.bootstrap4.min.css">
  <script src="/assets/js/main.js"></script>
`)

	_buffer.WriteString(`
  <title>`)
	_buffer.WriteString(`mypage`)

	_buffer.WriteString(`</title>
</head>
<body class="d-flex flex-column" style="min-height: 100vh">
  <nav class="navbar navbar-toggleable-md navbar-expand-lg navbar-light static-top v-navbar">
    <button class="navbar-toggler navbar-toggler-right" type="button" data-toggle="collapse" data-target="#navbarResponsive" aria-controls="navbarResponsive" aria-expanded="false" aria-label="Toggle navigation">
      <i class="fa fa-bars"></i>
    </button>
    <a class="navbar-brand" href="/">Vegeta</a>
    <div id="navbarResponsive" class="collapse navbar-collapse">
      <ul class="navbar-nav mr-auto">
        <li class="nav-item"><a class="nav-link" href="/contact">問い合わせ</a></li>
      </ul>
      <ul class="navbar-nav">
        `)
	if args.IsAuthed() {
		_buffer.WriteString(`
          <li class="nav-item dropdown">
            <a class="nav-link dropdown-toggle dropdown-toggle-split" href="" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false"><i class="fa fa-user" aria-hidden="true"></i> ユーザー</a>
            <div class="dropdown-menu">
              <a class="dropdown-item" href="/mypage"><i class="fa fa-pagelines" aria-hidden="true"></i> 観察</a>
              <div class="dropdown-divider"></div>
              `)
		if args.IsAdmin() {
			_buffer.WriteString(`
                <a class="dropdown-item" href="/mypage/admin"><i class="fa fa-lock" aria-hidden="true"></i> ユーザー管理パネル</a>
              `)
		} else {
			_buffer.WriteString(`
                <a class="dropdown-item" href="/mypage/settings"><i class="fa fa-cog" aria-hidden="true"></i> 設定</a>
              `)
		}
		_buffer.WriteString(`
            </div>
          </li>
          <li class="nav-item">
            <a class="nav-link" href="/mypage/logout"><i class="fa fa-sign-out" aria-hidden="true"></i> ログアウト</a>
          </li>
        `)
	} else {
		_buffer.WriteString(`
          <li class="nav-item">
            <a class="nav-link" href="/login"><i class="fa fa-sign-in" aria-hidden="true"></i> ログイン</a>
          </li>
        `)
	}
	_buffer.WriteString(`
      </ul>
    </div>
  </nav>
  <main class="mb-auto">
    `)

	_buffer.WriteString(`
  </main>
  <footer class="footer">
    <p>© `)
	hero.FormatInt(int64(args.Year()), _buffer)
	_buffer.WriteString(` <a class="text-white" href="https://twitter.com/CodeHex">CodeHex</a></p>
  </footer>
  `)
	_buffer.WriteString(`
  <script src="/assets/js/admin.js"></script>
`)

	_buffer.WriteString(`
</body>
</html>`)
	w.Write(_buffer.Bytes())

}
