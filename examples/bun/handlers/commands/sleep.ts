import {
    CommandHandler,
    dateToTimestamp,
    getOption,
    snowflakeToDate,
} from "../common.ts";
import { RESTPostAPIWebhookWithTokenJSONBody } from "discord-api-types/v10";

export const sleepCommandHandler: CommandHandler = async (interaction) => {
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
};
