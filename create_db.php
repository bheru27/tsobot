<?php
unlink(".cache/tsobot.db");
$sql = new SQLite3(".cache/tsobot.db");
$sql->query("
CREATE TABLE `clickbait` (
`id` INTEGER PRIMARY KEY AUTOINCREMENT,
`hash` TEXT NOT NULL,
`url` TEXT NOT NULL,
`title` TEXT NOT NULL,
`createdat` INTEGER NOT NULL
);

CREATE TABLE `log` (
`id` INTEGER PRIMARY KEY AUTOINCREMENT,
`chan` TEXT NOT NULL,
`nick` TEXT NOT NULL,
`line` TEXT NOT NULL,
`time` INTEGER NOT NULL
);
");

