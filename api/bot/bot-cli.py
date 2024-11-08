import discord
from discord.ext import commands
import requests
import os
import urllib.parse
from dotenv import load_dotenv
from discord.ext.commands import DefaultHelpCommand
import re
import hashlib
from requests.exceptions import SSLError, HTTPError, RequestException

load_dotenv()

DISCORD_TOKEN = os.getenv('DISCORD_TOKEN') or ""
RETRIEVAL_API_URL = os.getenv('RETRIEVAL_API_URL')
CHATBOT_API_URL = os.getenv('CHATBOT_API_URL')
AUTH_BEARER = os.getenv('AUTH_BEARER')
UPGRADE_API_URL = os.getenv('UPGRADE_API_URL')

bot = commands.Bot(command_prefix='!', intents=discord.Intents.all())

class CustomHelpCommand(DefaultHelpCommand):
    async def send_bot_help(self, mapping):
        embed = discord.Embed(title="Help - Available Commands", color=discord.Color.blue())
        for cog, commands in mapping.items():
            for command in commands:
                embed.add_field(name=f"!{command.name}", value=command.help, inline=False)
        await self.get_destination().send(embed=embed)

bot.help_command = CustomHelpCommand()

@bot.event
async def on_ready():
    print(f'{bot.user} has connected to Discord!')

def hash_user_id(user_id):
    return hashlib.sha256(user_id.encode()).hexdigest()

def chunk_text(text, max_chars=1024):
    paragraphs = text.split('\n')
    chunks = []
    current_chunk = ""
    current_char_count = 0

    for paragraph in paragraphs:
        if current_char_count + len(paragraph) + 1 > max_chars:
            chunks.append(current_chunk)
            current_chunk = paragraph
            current_char_count = len(paragraph) + 1
        else:
            if current_chunk:
                current_chunk += "\n" + paragraph
            else:
                current_chunk = paragraph
            current_char_count += len(paragraph) + 1

    if current_chunk:
        chunks.append(current_chunk)

    return chunks

async def send_embed(ctx, title, content):
    chunks = chunk_text(content)
    for chunk in chunks:
        embed = discord.Embed(title=title, description=chunk, color=discord.Color.blue())
        await ctx.send(embed=embed)

async def get_user_info(ctx):
    user_id = str(ctx.author.id)
    hashed_id = hash_user_id(user_id)
    try:
        response = requests.get(f"{UPGRADE_API_URL}/get-user-info", params={"id_hash": hashed_id})
        response.raise_for_status()
        return response.json()
    except (HTTPError, RequestException) as e:
        await ctx.send("Error retrieving user info. Please try again later.")
        print(f"Error: {e}")
        return None
    except SSLError as e:
        await ctx.send("SSL error encountered. Please contact support.")
        print(f"SSL Error: {e}")
        return None

async def enforce_credits(ctx):
    user_info = await get_user_info(ctx)
    if not user_info:
        return False, None
    user_id = str(ctx.author.id)
    hashed_id = hash_user_id(user_id)
    try:
        response = requests.post(f"{UPGRADE_API_URL}/enforce-credits", json={"id_hash": hashed_id}, headers={"Content-Type": "application/json"})
        response.raise_for_status()
        if response.status_code == 402:
            await ctx.send("You have run out of credits. Please upgrade your membership.")
            return False, None
        remaining_credits = response.json().get('credits', None)
        return True, remaining_credits
    except (HTTPError, RequestException) as e:
        await ctx.send("Error enforcing credits. Please try again later.")
        print(f"Error: {e}")
        return False, None
    except SSLError as e:
        await ctx.send("SSL error encountered. Please contact support.")
        print(f"SSL Error: {e}")
        return False, None

async def notify_credits(ctx, remaining_credits):
    if remaining_credits is not None:
        await send_embed(ctx, "", f"You have {remaining_credits} credits remaining.")

@bot.command(name='ask', help='Ask an investment research question')
async def ask(ctx: commands.Context, *, question: str) -> None:
    success, remaining_credits = await enforce_credits(ctx)
    if not success:
        return
    try:
        question = urllib.parse.quote(question)
        headers = {"Authorization": f"Bearer {AUTH_BEARER}", "Content-Type": "application/json"}
        response = requests.post(f"{CHATBOT_API_URL}/chat?prompt={question}", headers=headers)
        response.raise_for_status()
        answer = response.text
        await send_embed(ctx, "Answer to your question", answer)
    except (HTTPError, RequestException) as e:
        await ctx.send("Error processing your question. Please try again later.")
        print(f"Error: {e}")
    except SSLError as e:
        await ctx.send("SSL error encountered. Please contact support.")
        print(f"SSL Error: {e}")
    await notify_credits(ctx, remaining_credits)
