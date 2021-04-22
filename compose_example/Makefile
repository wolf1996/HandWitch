build_base_image:
	docker build ../ -t glazedcurd/handwitch

build_telegram_image:
	docker build ./ -f tg.Dockerfile -t glazedcurd/handwitchtelegram

build_slack_image:
	docker build ./ -f slack.Dockerfile -t glazedcurd/handwitchslack

compose_up:
	docker-compose -f docker-compose.dev.yml up --build

rebuild_base: 

all: build_base_image build_telegram_image build_slack_image compose_up