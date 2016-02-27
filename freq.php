<?php
$words = file_get_contents('full-emoji-list.txt');
$freq = [];
foreach(preg_split('/\s+/', $words) as $word) {
    $freq[$word]++;
}
arsort($freq);
var_dump($freq);
