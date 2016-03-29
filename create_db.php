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
);");

