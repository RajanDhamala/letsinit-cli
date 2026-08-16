#!/usr/bin/env node
import inquirer from "inquirer";
import fs from "fs-extra";
import path from "path";
import { exec } from "child_process";
import chalk from "chalk";
import { fileURLToPath } from "url";
import { createSpinner } from "nanospinner";

const args = process.argv.slice(2);
const isVerbose = args.includes("-v") || args.includes("--verbose");

function log(...x) {
  if (isVerbose) console.log(...x);
}

const cliRoot = path.dirname(fileURLToPath(import.meta.url));

const templateDirectories = {
  Js_react: "react-js",
  Ts_react: "react-ts",
  Nextjs: "nextjs",
  Express_Prisma: "express-prisma",
  Express_Mongoose: "express-mongoose",
  Fiber_Sqlc: "fiber-sqlc",
  Django: "django",
  FastAPI: "fastapi",
  Expo: "expo",
  Go_http: "go-http",
};


const nodeStacks = new Set([
  "Js_react",
  "Ts_react",
  "Nextjs",
  "Express_Prisma",
  "Express_Mongoose",
  "Expo",
]);

const pythonStacks = new Set(["Django", "FastAPI"]);

const systemPythonCommand = process.platform === "win32" ? "py -3" : "python3";
const virtualEnvironmentPython = process.platform === "win32"
  ? ".venv\\Scripts\\python.exe"
  : ".venv/bin/python";

const ecosystemInstallCommands = {
  Fiber_Sqlc: "go mod download",
};

const nodeStartScripts = {
  Js_react: "dev",
  Ts_react: "dev",
  Nextjs: "dev",
  Express_Prisma: "dev",
  Express_Mongoose: "dev",
  Expo: "start",
};

// Async runCommand for proper spinner animation
const runCommand = (command, options = {}, message) => {
  return new Promise((resolve, reject) => {
    const spinner = message ? createSpinner(message).start() : null;
    const child = exec(command, options, (err, stdout, _stderr) => {
      if (err) {
        if (spinner) spinner.error({ text: message || "Command failed" });
        reject(err);
      } else {
        if (spinner) spinner.success({ text: message });
        resolve(stdout);
      }
    });

    // Show output only in verbose mode
    if (isVerbose) {
      child.stdout.pipe(process.stdout);
      child.stderr.pipe(process.stderr);
    }
  });
};

const Main = async () => {
  console.log(chalk.cyan.bold("\n LetsInit Setup Starting...\n"));
  const stackChoices = [
    { name: "React + JS", value: "Js_react" },
    { name: "React + TS", value: "Ts_react" },
    { name: "Next.js + TS", value: "Nextjs" },
    { name: "Express + Prisma", value: "Express_Prisma" },
    { name: "Express + Mongoose", value: "Express_Mongoose" },
    { name: "Fiber + SQLC", value: "Fiber_Sqlc" },
    { name: "Go+SQLC", value: "Go_http" },
    { name: "Django", value: "Django" },
    { name: "FastAPI", value: "FastAPI" },
    { name: "Expo", value: "Expo" },
  ];
  const { stack_option } = await inquirer.prompt([
    {
      name: "stack_option",
      type: "list",
      message: "Choose your stack:",
      choices: stackChoices,
    },
  ]);
  log(chalk.gray("Stack chosen:"), stack_option);

  let package_manager = null;
  if (nodeStacks.has(stack_option)) {
    const answer = await inquirer.prompt([
      {
        name: "package_manager",
        type: "list",
        message: "Choose your package manager:",
        choices: [
          { name: "npm", value: "npm" },
          { name: "pnpm", value: "pnpm" },
        ],
      },
    ]);
    package_manager = answer.package_manager;
    log(chalk.gray("Package manager:"), package_manager);
  }

  let create_python_venv = false;
  if (pythonStacks.has(stack_option)) {
    const answer = await inquirer.prompt([
      {
        name: "create_python_venv",
        type: "confirm",
        message: "Create a Python virtual environment?",
        default: true,
      },
    ]);
    create_python_venv = answer.create_python_venv;
    log(chalk.gray("Create Python virtual environment:"), create_python_venv);
  }

  let folder_name = "letsinit-project";

  const { folder_name: fn } = await inquirer.prompt([
    {
      name: "folder_name",
      type: "input",
      message: "Enter your project folder name:",
      default: folder_name,
    },
  ]);
  folder_name = fn;
  log(chalk.gray("Folder name:"), folder_name);

  const targetDir = path.resolve(folder_name);
  const templateDir = path.join(
    cliRoot,
    "templates",
    templateDirectories[stack_option]
  );

  if (!fs.existsSync(templateDir)) {
    throw new Error(`Template not found for ${stack_option}`);
  }

  if (fs.existsSync(targetDir) && fs.readdirSync(targetDir).length > 0) {
    throw new Error(`Folder "${folder_name}" already exists and is not empty.`);
  }

  const templateSpinner = createSpinner(
    " Creating your project from a template..."
  ).start();

  try {
    fs.copySync(templateDir, targetDir, { overwrite: false });

    const templateFilesToRename = [
      ["_gitignore", ".gitignore"],
      [".env.example", ".env"],
    ];

    for (const [sourceName, targetName] of templateFilesToRename) {
      const sourcePath = path.join(targetDir, sourceName);
      if (fs.existsSync(sourcePath)) {
        fs.moveSync(sourcePath, path.join(targetDir, targetName));
      }
    }

    templateSpinner.success({ text: " Project brewed successfully!" });
  } catch (err) {
    templateSpinner.error({ text: " Failed to brew project." });
    throw err;
  }

  if (package_manager === "pnpm") {
    fs.removeSync(path.join(targetDir, "package-lock.json"));
  } else if (package_manager === "npm") {
    fs.removeSync(path.join(targetDir, "pnpm-lock.yaml"));
  }

  if (create_python_venv) {
    await runCommand(
      `${systemPythonCommand} -m venv .venv`,
      { cwd: targetDir },
      " Creating Python virtual environment..."
    );

    await runCommand(
      `${virtualEnvironmentPython} -m pip install -r requirements.txt`,
      { cwd: targetDir },
      " Installing Python dependencies..."
    );
  } else {
    const installCommand = package_manager
      ? `${package_manager} install`
      : ecosystemInstallCommands[stack_option];

    if (installCommand) {
      await runCommand(
        installCommand,
        { cwd: targetDir },
        package_manager
          ? ` Installing dependencies with ${package_manager}...`
          : " Installing dependencies..."
      );
    }
  }

  const ignorePath = path.join(targetDir, ".gitignore");
  if (!fs.existsSync(ignorePath)) fs.writeFileSync(ignorePath, ".env\n", "utf-8");
  else {
    const content = fs.readFileSync(ignorePath, "utf-8");
    if (!content.includes(".env")) fs.appendFileSync(ignorePath, "\n.env\n");
  }

  if (stack_option === "Express_Prisma") {
    const prismaGenerateCommand = package_manager === "pnpm"
      ? "pnpm exec prisma generate"
      : "npm exec -- prisma generate";

    await runCommand(
      prismaGenerateCommand,
      { cwd: targetDir },
      " Generating Prisma Client..."
    );
  }

  console.log(chalk.green.bold("Project setup completed!"));
  console.log(chalk.cyanBright(`→ cd ${JSON.stringify(folder_name)}`));

  const pythonCommand = create_python_venv
    ? virtualEnvironmentPython
    : systemPythonCommand;

  if (pythonStacks.has(stack_option) && !create_python_venv) {
    console.log(
      chalk.cyanBright(`→ ${pythonCommand} -m pip install -r requirements.txt`)
    );
  }

  if (nodeStacks.has(stack_option)) {
    console.log(
      chalk.cyanBright(
        `→ ${package_manager} run ${nodeStartScripts[stack_option]}`
      )
    );
  } else if (stack_option == "Go_http") {
    console.log(chalk.cyanBright("→ go run ./cmd/api/main.go"));
  } else if (stack_option == "Fiber_Sqlc") {
    console.log(chalk.cyanBright("→ go run main.go"));
  } else if (stack_option === "Django") {
    console.log(chalk.cyanBright(`→ ${pythonCommand} manage.py migrate`));
    console.log(chalk.cyanBright(`→ ${pythonCommand} manage.py runserver`));
  } else {
    console.log(chalk.cyanBright(`→ ${pythonCommand} main.py`));
  }
};

Main().catch((e) => {
  console.error(chalk.red("Error during setup:"), e.message);
  process.exit(1);
});
