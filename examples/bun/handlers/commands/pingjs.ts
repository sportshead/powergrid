import { CommandHandler, getOption } from "../common.ts";
import {
    APIInteractionResponse,
    InteractionResponseType,
} from "discord-api-types/v10";

export const pingjsCommandHandler: CommandHandler = async (interaction) => {
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
};
