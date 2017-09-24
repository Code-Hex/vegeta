var path = require('path');
var fs = require('fs');
var pkg = JSON.parse(fs.readFileSync('./package.json'));
var assetsPath = path.resolve(pkg.path.assetsDir);

var gulp = require('gulp');

// sass compiler
var sass = require('gulp-sass');

// add vender prifix
var autoprefixer = require('gulp-autoprefixer');

// error handling
var plumber = require('gulp-plumber');

// for sass
var bourbon = require("node-bourbon").includePaths;
var neat = require("node-neat").includePaths;

gulp.task('sass', function() {
  gulp.src(path.join(assetsPath, 'sass/main.scss'))
  .pipe(plumber())
  .pipe(sass({
    includePaths: bourbon,
    includePaths: neat
  }))
  .pipe(autoprefixer())
  .pipe(gulp.dest(path.join(assetsPath, 'css/')));
});

gulp.task("browserSync", function() {
  browserSync({
    server: {
      baseDir: assetsPath
    }
  })
});

// If you run `gulp` command, it is monioring sass files.
// Invoke auto compile after sass files are changed.
gulp.task('watch', ["browserSync", "sass"], function() {
  gulp.watch(path.join(assetsPath, 'sass/**/*.scss'),['sass']);
});