import {
    APIApplicationCommandInteractionDataStringOption,
    APIChatInputApplicationCommandInteraction,
    APIInteraction,
    APIInteractionResponse,
    InteractionResponseType,
    InteractionType,
} from "discord-api-types/v10";


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
        if (!checkInteraction(interaction)) {
            console.error(
                `[${new Date().toUTCString()}] Unknown interaction type`,
                interaction,
            );
            return new Response("unknown interaction", { status: 400 });
        }

        const name =
            (
                interaction.data.options?.find((o) => o.name === "name") as
                    | APIApplicationCommandInteractionDataStringOption
                    | undefined
            )?.value ?? "world";

        const res: APIInteractionResponse = {
            type: InteractionResponseType.ChannelMessageWithSource,
            data: {
                content: `Hello ${name} from Bun version ${Bun.version}!\nHostname: \`${process.env.HOSTNAME}\``,
            },
        };

        console.log(
            `[${new Date().toUTCString()}] Responding to interaction ${
                interaction.id
            } from @${interaction.user?.username} (${interaction.user?.id}`,
            res,
        );

        return new Response(JSON.stringify(res), {
            status: 200,
            headers: { "Content-Type": "application/json" },
        });
    },
    port: 3000,
});

const checkInteraction = (
    interaction: APIInteraction,
): interaction is APIChatInputApplicationCommandInteraction =>
    interaction?.type === InteractionType.ApplicationCommand &&
    interaction?.data?.name === "pingjs";
