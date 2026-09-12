Siga todas as instruções do AGENTS.md do repositório.

Leia integralmente docs/project/ai/review-policy.md antes de iniciar a revisão.

Estabeleça o objeto da revisão

Identifique:

a unidade ECOM-_._;

o modo pre-commit ou pull-request;

a branch, worktree, commit ou pull request revisado;

a versão aprovada da implementation note;

as ações autorizadas pelo pedido atual.

Se o objeto não puder ser determinado sem risco de revisar o estado errado, pare
e peça esclarecimento.

Preserve a independência

Reconstrua o contrato pelas fontes primárias do repositório. O relatório do
implementador é contexto auxiliar, não evidência.

Comece em modo somente leitura. Não modifique arquivos, aplique correções, faça
stage, commit, push, comentário remoto, aprovação de pull request ou merge sem
autorização explícita posterior.

Revise

Compare a unidade e sua implementation note com:

o diff completo, incluindo arquivos não rastreados quando o modo for
pre-commit;

todos os consumidores dos contratos alterados;

testes e evidência de regressão;

contratos públicos, migrations e comportamento persistente afetado;

documentação do comportamento atual e execution status;

resultados reais dos gates aplicáveis.

Use os critérios, classificações e formato de relatório definidos em
docs/project/ai/review-policy.md.

Não invente achados. Não aceite afirmações sem evidência. Não trate um gate
ignorado, indisponível ou não executado como aprovado.

Conclua

Retorne exatamente um veredito:

approved;

approved with non-blocking observations;

changes required;

blocked by missing evidence.

Apresente os achados antes do resumo. Confirme o estado final do worktree e diga
se houve qualquer ação de escrita. Aprovação de revisão não autoriza commit,
push, merge ou deploy.
