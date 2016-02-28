package main

var emoji map[string]string = map[string]string{
<?php
$other = [];
$emoji = file_get_contents('full-emoji-list.html');
preg_match_all("/<tr>(.*?)<\/tr>/mis", $emoji, $m, PREG_SET_ORDER);
array_shift($m);
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
        println($unicode, trim($name));
    }
//    echo $unicode, ' ', strtolower($name), ' ', implode(' ', $words), "\n";
}

function println($unicode, $name) {
    echo '"', str_replace(' ', '_', strtolower($name)), '": "', $unicode, '",', "\n";
}
?>
}

var other map[string][]string = map[string][]string{
<?php
foreach($other as $word => $emojis) {
    echo '"', str_replace(' ', '_', strtolower($word)), '": []string{', 
        '"',
        implode('","', $emojis),
        '"},', "\n";
}
?>
}
