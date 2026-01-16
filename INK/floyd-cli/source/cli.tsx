#!/usr/bin/env node
import React from 'react';
import {render} from 'ink';
import meow from 'meow';
import App from './app.js';

const cli = meow(
	`
	Usage
	  $ floyd-cli

	Options
		--name  Your name

	Examples
	  $ floyd-cli --name=Jane
	  Hello, Jane
`,
	{
		importMeta: import.meta,
		flags: {
			name: {
				type: 'string',
			},
			chrome: {
				type: 'boolean',
			},
		},
	},
);

render(<App name={cli.flags.name} chrome={cli.flags.chrome} />);
