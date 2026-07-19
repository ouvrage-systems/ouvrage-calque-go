@echo off
:: @ocq:Config(comment_style="::")

:: @ocq:Replace(with="SET COMPILER_PATH=C:\Program Files\MSBuild\bin", when=(env.ENV_NAME == "production"))
SET COMPILER_PATH=C:\DevCompiler\bin

echo Running build using %COMPILER_PATH%...
