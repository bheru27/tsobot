package main

var emoji map[string]string = map[string]string{
<?php
$other = [];
$emoji = file_get_contents('full-emoji-list.html');
preg_match_all("/<tr>(.*?)<\/tr>/mis", $emoji, $m, PREG_SET_ORDER);
array_shift($m);
$names = [];
foreach($m as $sample) {
    if(!preg_match("/<td class='chars'>(.*)<\/td>/", $sample[1], $u)) continue;
    $unicode = $u[1];
    if(!preg_match_all("/<td class='name'>(.*)<\/td>/", $sample[1], $n)) continue;
    if(preg_match_all("/<a.*?>(.*?)<\/a>/", $n[1][1], $o)) {
        $words = $o[1];
        foreach($words as $word) {
            $other[$word][] = $unicode;
        }
    }
    foreach(explode("<br>≊", $n[1][0]) as $name) {
        $name = sanitize($name);
        if(strstr($name, '_type_') || in_array($name, $names)) continue;
        println($unicode, $name);
        $names[] = $name;
    }
}

function println($unicode, $name) {
    echo '"', $name, '": "', $unicode, '",', "\n";
}

function sanitize($name) {
    return preg_replace('/[^\w\+]+/', '_', trim(strtolower($name)));
}
?>
}

var other map[string][]string = map[string][]string{
<?php
foreach($other as $word => $emojis) {
    echo '"', sanitize($word), '": []string{', 
        '"',
        implode('","', $emojis),
        '"},', "\n";
}
?>
}

var jmote map[string][]string = map[string][]string{
<?php
$d = dir('emot');
while(($f = $d->read())!==false) {
    if(is_dir($f) || !preg_match('/^(\w+)\.txt$/', $f, $m)) continue;
    echo '"', sanitize($m[1]), '": []string{"', implode('","', array_map('trim', file("emot/$f"))), '"},', "\n";
}
?>
}
