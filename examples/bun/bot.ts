import { APIInteraction } from "discord-api-types/v10";
import { commandHandlers, isApplicationCommand } from "./handlers/command.ts";
import {
    autocompleteHandlers,
    isAutocomplete,
} from "./handlers/autocomplete.ts";

Bun.serve({
    async fetch(req: Request) {
        // we can assume the request has been verified already
        const url = new URL(req.url);
        if (url.pathname === "/healthz") {
            return new Response("ok", { status: 200 });
        }
        if (url.pathname !== "/") {
            return new Response("not found", { status: 404 });
        }
        if (req.method !== "POST") {
            return new Response("invalid method", {
                status: 405,
                headers: {
                    Allow: "POST",
                },
            });
        }

        const interaction = (await req.json()) as APIInteraction;
        if (isApplicationCommand(interaction)) {
            return commandHandlers[interaction.data.name](interaction);
        }
        if (isAutocomplete(interaction)) {
            return autocompleteHandlers[interaction.data.name](interaction);
        }

        console.error(
            `[${new Date().toUTCString()}] Unknown interaction type`,
            interaction,
        );
        return new Response("unknown interaction", { status: 400 });
    },
    port: 3000,
});
