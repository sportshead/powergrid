import {
    APIApplicationCommandInteractionDataStringOption,
    APIChatInputApplicationCommandInteraction,
    APIInteraction,
    APIInteractionResponse,
    InteractionResponseType,
    InteractionType,
    RESTPostAPIWebhookWithTokenJSONBody,
} from "discord-api-types/v10";

const getOption = <T>(
    interaction: APIChatInputApplicationCommandInteraction,
    name: string,
    defaultValue: T,
) =>
    (
        interaction.data.options?.find((o) => o.name === name) as
            | { value: T | undefined }
            | undefined
    )?.value ?? defaultValue;

// https://discord.com/developers/docs/reference#snowflakes-snowflake-id-format-structure-left-to-right
const DISCORD_EPOCH = 1420070400000;
const snowflakeToDate = (snowflake: number) =>
    new Date(Number(BigInt(snowflake) >> 22n) + DISCORD_EPOCH);

const dateToTimestamp = (d: Date) => Math.floor(d.getTime() / 1000);

const commandHandlers: Record<
    string,
    (
        interaction: APIChatInputApplicationCommandInteraction,
    ) => Promise<Response>
> = {
    pingjs: async (interaction) => {
        const name = getOption<string>(interaction, "name", "world");

        const res: APIInteractionResponse = {
            type: InteractionResponseType.ChannelMessageWithSource,
            data: {
                content: `Hello ${name} from Bun version ${Bun.version}!\nHostname: \`${process.env.HOSTNAME}\``,
            },
        };

        console.log(
            `[${new Date().toUTCString()}] Responding to interaction ${
                interaction.id
            } from @${interaction.member?.user?.username} (${interaction.member
                ?.user?.id}`,
            res,
        );

        return new Response(JSON.stringify(res), {
            status: 200,
            headers: { "Content-Type": "application/json" },
        });
    },
    sleep: async (interaction) => {
        const time = getOption<number>(interaction, "time", 5000);
        await Bun.sleep(time);

        const req: RESTPostAPIWebhookWithTokenJSONBody = {
            content: `zzzzz...\nSlept for ${time}ms, from <t:${dateToTimestamp(
                snowflakeToDate(parseInt(interaction.id)),
            )}:T> to <t:${dateToTimestamp(new Date())}:T>`,
        };

        console.log(
            `[${new Date().toUTCString()}] Responding to deferred interaction ${
                interaction.id
            } from @${interaction.member?.user?.username} (${interaction.member
                ?.user?.id}`,
            req,
        );

        await fetch(
            `https://discord.com/api/webhooks/${interaction.application_id}/${interaction.token}`,
            {
                method: "POST",
                body: JSON.stringify(req),
                headers: {
                    "Content-Type": "application/json",
                },
            },
        );

        return new Response("ok", { status: 200 });
    },
};

const checkInteraction = (
    interaction: APIInteraction,
): interaction is APIChatInputApplicationCommandInteraction =>
    interaction?.type === InteractionType.ApplicationCommand &&
    interaction?.data?.name in commandHandlers;

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

        return commandHandlers[interaction.data.name](interaction);
    },
    port: 3000,
});
