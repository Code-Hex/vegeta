"use strict";

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

// sass concat
var concat = require('gulp-concat');

// for sass
var bourbon = require("node-bourbon");
var neat = require("node-neat");

// minify css
var minify = require('gulp-minify-css');

// merge
var merge = require('merge-stream');

gulp.task('sass', function() {
  var scssStream = gulp.src('sass/*.scss')
    .pipe(concat('concat.scss'))
    .pipe(plumber())
    .pipe(sass({
      includePaths: [
        bourbon.includePaths,
        neat.includePaths
      ]
    }))
    .pipe(autoprefixer());
  
  merge(scssStream)
    .pipe(concat('main.css'))
    .pipe(minify())
    .pipe(gulp.dest(path.join(assetsPath, 'css/')));
});

// If you run `gulp` command, it is monioring sass files.
// Invoke auto compile after sass files are changed.
gulp.task('watch', ["sass"], function() {
  gulp.watch('sass/*.scss', ['sass']);
});