<?php
$emoji = file_get_contents('full-emoji-list.html');
preg_match_all("/<tr>(.*?)<\/tr>/mis", $emoji, $m, PREG_SET_ORDER);
array_shift($m);
foreach($m as $sample) {
    if(!preg_match("/<td class='chars'>(.*)<\/td>/", $sample[1], $u)) continue;
    $unicode = $u[1];
    if(!preg_match_all("/<td class='name'>(.*)<\/td>/", $sample[1], $n)) continue;
    $name = $n[1][0];
/*    preg_match_all("/<a.*?>(.*?)<\/a>/", $n[1][1], $o);
    $words = $o[1];*/

//    echo $unicode, ' ', strtolower($name), ' ', implode(' ', $words), "\n";
    echo '"', str_replace(' ', '_', strtolower($name)), '": "', $unicode, '",', "\n";
}

    

