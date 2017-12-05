"use strict";

var path = require('path');
var fs = require('fs');
var pkg = JSON.parse(fs.readFileSync('./package.json'));
var assetsPath = path.resolve(pkg.path.assetsDir);

var gulp = require('gulp');

// browserify
var browserify = require('gulp-browserify');

// typescript
var ts = require('gulp-typescript');
var tsProject = ts.createProject('tsconfig.json');

// javascript minify
var uglify = require('gulp-uglify');

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

gulp.task('ts', function() {
  tsProject.src()
  .pipe(tsProject())
  .pipe(browserify({jquery:'jquery-browserify'}))
  .pipe(uglify())
  .pipe(gulp.dest(path.join(assetsPath, 'js/')));
})

gulp.task('js', function() {
  gulp.src('js/*.js')
    .pipe(plumber())
    .pipe(browserify({jquery:'jquery-browserify'}))
    .pipe(uglify())
    .pipe(gulp.dest(path.join(assetsPath, 'js/')));
})

gulp.task('css', function() {
  gulp.src([
    'node_modules/c3/c3.min.css',
    'node_modules/flatpickr/dist/flatpickr.min.css'
  ])
  .pipe(plumber())
  .pipe(gulp.dest(path.join(assetsPath, 'css/')));
})

gulp.task('bootstrap', function() {
  gulp.src('node_modules/bootstrap/scss/bootstrap.scss')
  .pipe(plumber())
  .pipe(sass())
  .pipe(minify())
  .pipe(gulp.dest(path.join(assetsPath, 'css/')));

  gulp.src([
    'node_modules/bootstrap/dist/js/bootstrap.min.js',
    'node_modules/jquery/dist/jquery.min.js',
    'node_modules/tether/dist/js/tether.min.js'
  ])
  .pipe(uglify())
  .pipe(gulp.dest(path.join(assetsPath, 'js/')));
})

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
})

// If you run `gulp` command, it is monioring sass files.
// Invoke auto compile after sass files are changed.
gulp.task('watch', ["sass"], function() {
  gulp.watch('sass/*.scss', ['sass']);
})